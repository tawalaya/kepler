/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package source

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

const (
	scaphandrePackageNamePathTemplate = "/var/scaphandre/intel-rapl:%d/"
)

var (
	scaphandreEventPaths map[int]string
)

func init() {
	scaphandreEventPaths = map[int]string{}
	for i := 0; i < numPackages; i++ {
		packagePath := fmt.Sprintf(scaphandrePackageNamePathTemplate+energyFile, i)
		_, err := os.ReadFile(packagePath)
		if err == nil {
			scaphandreEventPaths[i] = packagePath
		}
	}
}

type PowerScaphandre struct{}

func (PowerScaphandre) GetName() string {
	return "rapl-sysfs-scaphandre"
}

func readScaphandreEnergy() map[int]uint64 {
	energy := map[int]uint64{}
	var err error
	var e uint64
	var data []byte
	for pkID, path := range scaphandreEventPaths {
		if data, err = os.ReadFile(path); err != nil {
			klog.V(3).Infoln(err)
			continue
		}
		if e, err = strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64); err != nil {
			klog.V(3).Infoln(err)
			continue
		}
		e /= 1000 /*mJ*/
		energy[pkID] = e
	}
	return energy
}

// getEnergy returns the sum of the energy consumption of all sockets for a given event
func getScaphandreEnergy() (uint64, error) {
	energy := uint64(0)
	energyMap := readScaphandreEnergy()
	for _, e := range energyMap {
		energy += e
	}
	return energy, nil
}

func (r *PowerScaphandre) IsSystemCollectionSupported() bool {
	path := fmt.Sprintf(scaphandrePackageNamePathTemplate, 0)
	_, err := os.ReadFile(path + energyFile)
	return err == nil
}

func (r *PowerScaphandre) GetAbsEnergyFromDram() (uint64, error) {
	return 0, fmt.Errorf("not supported")
}

func (r *PowerScaphandre) GetAbsEnergyFromCore() (uint64, error) {
	return 0, fmt.Errorf("not supported")
}

func (r *PowerScaphandre) GetAbsEnergyFromUncore() (uint64, error) {
	return getScaphandreEnergy()
}

func (r *PowerScaphandre) GetAbsEnergyFromPackage() (uint64, error) {
	return getScaphandreEnergy()
}

func (r *PowerScaphandre) GetAbsEnergyFromNodeComponents() map[int]NodeComponentsEnergy {
	packageEnergies := make(map[int]NodeComponentsEnergy)
	pkgEnergies := readScaphandreEnergy()
	for pkgID, pkgEnergy := range pkgEnergies {
		packageEnergies[pkgID] = NodeComponentsEnergy{
			Core:   pkgEnergy,
			DRAM:   0,
			Uncore: 0,
			Pkg:    pkgEnergy,
		}
	}

	return packageEnergies
}

func (r *PowerScaphandre) StopPower() {
}
