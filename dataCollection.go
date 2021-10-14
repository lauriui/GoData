package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Users struct which contains
// an array of users
type Arrays struct {
	Arrays []Array `json:"array"`
}

// User struct which contains a name
// a type and a list of social links
type Array struct {
	Name   string `json:"name"`
	Ip     string `json:"ip"`
	Site   string `json:"site"`
	Type   string `json:"type_arr"`
	Client string `json:"client"`
}

type Pools struct {
	Pools []Pool
}

type Pool struct {
	Id               string
	ArrayName        string
	PoolName         string
	Firmware         string
	Site             string
	Type             string
	Client           string
	PoolCapacity     float64
	PoolCapacityFree float64
	PoolCapacityUsed float64
	PoolCapacityPCT  float64
}

type Client struct {
	Name                  string
	StretchedP16Total     float64
	StretchedP16Free      float64
	StretchedP16MinLun    int
	StretchedZ141Total    float64
	StretchedZ141Free     float64
	StretchedZ141MinLun   int
	P16Total              float64
	P16Free               float64
	P16InternalTotal      float64
	P16InternalFree       float64
	P16InternalSSDTotal   float64
	P16InternalHDDTotal   float64
	P16InternalSSDFree    float64
	P16InternalHDDFree    float64
	P16InternalSSDMinLun  int
	P16InternalHDDMinLun  int
	P16ExternalTotal      float64
	P16ExternalFree       float64
	P16ExternalSSDTotal   float64
	P16ExternalHDDTotal   float64
	P16ExternalSSDFree    float64
	P16ExternalHDDFree    float64
	P16ExternalSSDMinLun  int
	P16ExternalHDDMinLun  int
	Z141Total             float64
	Z141Free              float64
	Z141InternalTotal     float64
	Z141InternalFree      float64
	Z141InternalSSDTotal  float64
	Z141InternalHDDTotal  float64
	Z141InternalSSDFree   float64
	Z141InternalHDDFree   float64
	Z141InternalSSDMinLun int
	Z141InternalHDDMinLun int
	Z141ExternalTotal     float64
	Z141ExternalFree      float64
	Z141ExternalSSDTotal  float64
	Z141ExternalHDDTotal  float64
	Z141ExternalSSDFree   float64
	Z141ExternalHDDFree   float64
	Z141ExternalSSDMinLun int
	Z141ExternalHDDMinLun int
	Total                 float64
	TotalFree             float64
}

func logError(Error string) {
	log_date := time.Now()
	years, month, day := log_date.Date()
	filename := "logs/test." + strconv.Itoa(years) + strconv.Itoa(int(month)) + strconv.Itoa(day) + ".log"
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)

	log.Println(Error)

}

func collectData(user, password, host, array, site, type_s, client_s, model string, test bool) (output Pools) {
	var client *ssh.Client
	var err error
	var poolData Pools
	client, err = connectToHostPW(user, password, host)
	if err != nil {
		client, err = connectToHostKB(user, password, host)
		if err != nil {
			errorString := "CollectData: ConnectToHostKB: " + array + ": " + err.Error()
			logError(errorString)
		} else {
			data, err := getData(client, model, array, test)
			if err != nil {
				logError(err.Error())
			}

			fw, err := getFw(client, model, array, test)
			if err != nil {
				logError(err.Error())
			}
			defer client.Close()

			poolData, err = parseData(data, fw, model, array, site, type_s, client_s)
			if err != nil {
				logError(err.Error())
			}
		}
	} else {
		data, err := getData(client, model, array, test)
		if err != nil {
			logError(err.Error())
		}

		fw, err := getFw(client, model, array, test)
		if err != nil {
			logError(err.Error())
		}
		defer client.Close()

		poolData, err = parseData(data, fw, model, array, site, type_s, client_s)
		if err != nil {
			logError(err.Error())
		}
	}

	return poolData
}

func getData(client *ssh.Client, model string, array string, test bool) ([]byte, error) {
	var output []byte
	var err error
	switch model {
	case "ibm":
		if test {
			err = nil
			output = []byte(`id,name,status,mdisk_count,vdisk_count,capacity,extent_size,free_capacity,virtual_capacity,used_capacity,real_capacity,overallocation,warning,easy_tier,easy_tier_status,compression_active,compression_virtual_capacity,compression_compressed_capacity,compression_uncompressed_capacity,parent_mdisk_grp_id,parent_mdisk_grp_name,child_mdisk_grp_count,child_mdisk_grp_capacity,type,encrypt,owner_type,site_id,site_name,data_reduction,used_capacity_before_reduction,used_capacity_after_reduction,overhead_capacity,deduplication_capacity_saving,reclaimable_capacity,easy_tier_fcm_over_allocation_max
			0,qwe4,online,14,50,123435046494208,1024,15360950534144,123422882781696,123430261094400,130849826856448,105,80,auto,balanced,no,0,0,0,0,Z141_SSD01,0,0,parent,yes,none,1,Z141,no,0,0,0,0,0,
			1,qwe3,online,14,61,12345046494208,1024,13348758355968,123489150040576,123421639157760,132858589771264,112,80,auto,balanced,no,0,0,0,1,P16_SSD01,0,0,parent,yes,none,2,P16,no,0,0,0,0,0,
			`)
		} else {
			session, err := client.NewSession()
			if err != nil {
				logError(err.Error())
			}
			output, err = session.CombinedOutput("lsmdiskgrp -bytes -delim ,")
		}
	case "huawei":
		if test {
			err = nil
			output = []byte(`
			ID  Name                  Disk Domain ID  Health Status  Running Status  Total Capacity  Free Capacity  Usage Type
			--  --------------------  --------------  -------------  --------------  --------------  -------------  ----------
			0   asd1         0               Normal         Online                123.410TB       123.616TB  LUN
			1   asd2         1               Normal         Online                123.257TB        123.465TB  LUN
			2   asd3  0               Normal         Online               123.121TB        123.088TB  LUN
			3   asd4  1               Normal         Online               123.748TB        123.231TB  LUN
			5   asd5         3               Normal         Online                123.886TB       123.378TB  LUN
			6   asd6         4               Normal         Online                123.886TB       123.878TB  LUN
		  `)
		} else {
			session, err := client.NewSession()
			if err != nil {
				logError(err.Error())
			}
			output, err = session.CombinedOutput("show storage_pool general")
		}
	case "3par":
		fmt.Println("3par")
	case "dell":
		fmt.Println("dell")
	}

	return output, err

}

func getFw(client *ssh.Client, model string, array string, test bool) ([]byte, error) {
	var output []byte
	var err error
	switch model {
	case "ibm":
		if test {
			err = nil
			output = []byte("code_level,8.3.1.5 (build 150.27.2104221539000)")
		} else {
			session, err := client.NewSession()
			if err != nil {
				logError(err.Error())
			}
			output, err = session.CombinedOutput("lssystem -delim ,| grep -i code")
		}
	case "huawei":
		if test {
			err = nil
			output = []byte(`
			System Name         : STRSQLZ1
			Health Status       : Normal
			Running Status      : Normal
			Total Capacity      : 610.723TB
			SN                  : 210235982610H3000008
			Location            : Z141_S5_14
			Product Model       : 6800 V3
			Product Version     : V300R006C20
			High Water Level(%) : 80
			Low Water Level(%)  : 20
			WWN                 : 210080d4a506b8ee
			Time                : 2021-10-09/12:16:06 UTC+03:00
			Patch Version       : SPH035`)
		} else {
			session, err := client.NewSession()
			if err != nil {
				logError(err.Error())
			}
			output, err = session.CombinedOutput("show system general")
		}
	case "3par":
		fmt.Println("3par")
		if test {
			fmt.Println(string("getting data"))
		} else {
			// session, err := client.NewSession()
			// if err != nil {
			// 	logError(err.Error())
			// }
			// output, err = session.CombinedOutput("ls -lrt /home/")
		}
	case "dell":
		fmt.Println("dell")
		if test {
			fmt.Println(string("getting data"))
		} else {
			// session, err := client.NewSession()
			// if err != nil {
			// 	logError(err.Error())
			// }
			// output, err = session.CombinedOutput("ls -lrt /home/")
		}
	}

	return output, err
}

func parseData(inputData []byte, inputFw []byte, model, array, site, type_s, client_s string) (output Pools, err error) {
	splitInputData := strings.Split(string(inputData), "\n")
	switch model {
	case "ibm":
		firmware := strings.Split(strings.Split(string(inputFw), ",")[1], " ")[0]
		for i := 1; i < len(splitInputData); i++ {
			lineSplit := strings.Split(splitInputData[i], ",")
			if len(lineSplit) > 1 && lineSplit[0] != "--" {
				var pool Pool
				pool.ArrayName = array
				pool.Id = strings.ReplaceAll(lineSplit[0], "	", "")
				pool.PoolName = lineSplit[1]
				pool.PoolCapacity, err = strconv.ParseFloat(lineSplit[5], 64)
				pool.PoolCapacityUsed, err = strconv.ParseFloat(lineSplit[9], 64)
				pool.PoolCapacityFree, err = strconv.ParseFloat(lineSplit[7], 64)
				pool.PoolCapacityPCT = pool.PoolCapacityUsed / pool.PoolCapacity
				pool.Firmware = firmware
				pool.Site = site
				pool.Type = type_s
				pool.Client = client_s
				output.Pools = append(output.Pools, pool)
			}
		}

	case "huawei":

		splitFW := strings.Split(string(inputFw), "\n")
		var pversion, patch, firmware string
		for index, line := range splitFW {
			if strings.Contains(line, "Product Version") {
				pversion = strings.ReplaceAll(strings.Split(line, ":")[1], " ", "")
			}
			if strings.Contains(line, "Patch Version") {
				patch = strings.ReplaceAll(strings.Split(line, ":")[1], " ", "")
			}
			index = index
		}
		firmware = pversion + ", " + patch
		// fmt.Println(firmware)
		for i := 3; i < len(splitInputData); i++ {

			line := strings.ReplaceAll(splitInputData[i], "	", "")
			re_leadclose_whtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
			re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
			line = re_leadclose_whtsp.ReplaceAllString(line, "")
			line = re_inside_whtsp.ReplaceAllString(line, " ")
			splitLine := strings.Split(line, " ")
			if len(splitLine) > 1 && splitLine[0] != "--" {
				var pool Pool
				pool.ArrayName = array
				pool.Id = splitLine[0]
				pool.PoolName = splitLine[1]
				var tcap float64
				if strings.Contains(splitLine[5], "PB") {
					tcap, err = strconv.ParseFloat(strings.ReplaceAll(splitLine[5], "PB", ""), 64)
					tcap = tcap * 1024 * 1024 * 1024 * 1024 * 1024
				} else if strings.Contains(splitLine[5], "TB") {
					tcap, err = strconv.ParseFloat(strings.ReplaceAll(splitLine[5], "TB", ""), 64)
					tcap = tcap * 1024 * 1024 * 1024 * 1024
				} else if strings.Contains(splitLine[5], "GB") {
					tcap, err = strconv.ParseFloat(strings.ReplaceAll(splitLine[5], "GB", ""), 64)
					tcap = tcap * 1024 * 1024 * 1024
				} else if strings.Contains(splitLine[5], "MB") {
					tcap, err = strconv.ParseFloat(strings.ReplaceAll(splitLine[5], "MB", ""), 64)
					tcap = tcap * 1024 * 1024
				} else if strings.Contains(splitLine[5], "KB") {
					tcap, err = strconv.ParseFloat(strings.ReplaceAll(splitLine[5], "KB", ""), 64)
					tcap = tcap * 1024
				} else if strings.Contains(splitLine[5], "B") {
					tcap, err = strconv.ParseFloat(strings.ReplaceAll(splitLine[5], "B", ""), 64)
				}
				pool.PoolCapacity = tcap
				var fcap float64
				if strings.Contains(splitLine[6], "PB") {
					fcap, err = strconv.ParseFloat(strings.ReplaceAll(splitLine[6], "PB", ""), 64)
					fcap = fcap * 1024 * 1024 * 1024 * 1024 * 1024
				} else if strings.Contains(splitLine[6], "TB") {
					fcap, err = strconv.ParseFloat(strings.ReplaceAll(splitLine[6], "TB", ""), 64)
					fcap = fcap * 1024 * 1024 * 1024 * 1024
				} else if strings.Contains(splitLine[6], "GB") {
					fcap, err = strconv.ParseFloat(strings.ReplaceAll(splitLine[6], "GB", ""), 64)
					fcap = fcap * 1024 * 1024 * 1024
				} else if strings.Contains(splitLine[6], "MB") {
					fcap, err = strconv.ParseFloat(strings.ReplaceAll(splitLine[6], "MB", ""), 64)
					fcap = fcap * 1024 * 1024
				} else if strings.Contains(splitLine[6], "KB") {
					fcap, err = strconv.ParseFloat(strings.ReplaceAll(splitLine[6], "KB", ""), 64)
					fcap = fcap * 1024
				} else if strings.Contains(splitLine[6], "B") {
					fcap, err = strconv.ParseFloat(strings.ReplaceAll(splitLine[6], "B", ""), 64)
				}
				pool.PoolCapacityFree = fcap
				pool.PoolCapacityUsed = pool.PoolCapacity - pool.PoolCapacityFree
				pool.PoolCapacityPCT = pool.PoolCapacityUsed / pool.PoolCapacity
				pool.Firmware = firmware
				pool.Site = site
				pool.Type = type_s
				pool.Client = client_s
				output.Pools = append(output.Pools, pool)
				// fmt.Println(pool)
			}
		}

	case "3par":
		fmt.Println("3par")

	case "dell":
		fmt.Println("dell")
	}

	return output, err
}

func connectToHostPW(user, password, host string) (*ssh.Client, error) {

	sshConfig := &ssh.ClientConfig{
		User:    user,
		Auth:    []ssh.AuthMethod{ssh.Password(password)},
		Timeout: 10 * time.Second,
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial("tcp", host+":22", sshConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func connectToHostKB(user, password, host string) (*ssh.Client, error) {

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
			answers := make([]string, len(questions))
			for i, _ := range answers {
				answers[i] = password
			}
			return answers, nil
		})},
		Timeout: 10 * time.Second,
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial("tcp", host+":22", sshConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func main() {
	username := "xxx"
	password := "xxx"
	test := false
	logError("Start")

	var jsonFileIBM *os.File
	var jsonFileHuawei *os.File
	var err error

	if test {
		jsonFileIBM, err = os.Open("test.json")
	} else {
		jsonFileIBM, err = os.Open("IBM.json")
	}

	if err != nil {
		logError(err.Error())
	}
	fmt.Println("Successfully Opened IBM.json")
	defer jsonFileIBM.Close()
	byteValueIBM, _ := ioutil.ReadAll(jsonFileIBM)
	var arraysIBM Arrays
	json.Unmarshal(byteValueIBM, &arraysIBM)
	if test {
		jsonFileHuawei, err = os.Open("test.json")
	} else {
		jsonFileHuawei, err = os.Open("huawei.json")
	}
	if err != nil {
		logError(err.Error())
	}
	fmt.Println("Successfully Opened huawei.json")
	defer jsonFileHuawei.Close()
	byteValueHuawei, _ := ioutil.ReadAll(jsonFileHuawei)
	var arraysHuawei Arrays
	json.Unmarshal(byteValueHuawei, &arraysHuawei)

	var pools Pools
	for i := 0; i < len(arraysIBM.Arrays); i++ {
		model := "ibm"
		if arraysIBM.Arrays[i].Client == "Telia" {
			logError("connecting to ibm host: " + arraysIBM.Arrays[i].Name)
			arrayPools := collectData(username, password, arraysIBM.Arrays[i].Ip, arraysIBM.Arrays[i].Name, arraysIBM.Arrays[i].Site, arraysIBM.Arrays[i].Type, arraysIBM.Arrays[i].Client, model, test)
			pools.Pools = append(pools.Pools, arrayPools.Pools...)
		}

	}

	for i := 0; i < len(arraysHuawei.Arrays); i++ {

		model := "huawei"
		if arraysHuawei.Arrays[i].Client == "Telia" {
			logError("connecting to huawei host : " + arraysHuawei.Arrays[i].Name)
			arrayPools := collectData(username, password, arraysHuawei.Arrays[i].Ip, arraysHuawei.Arrays[i].Name, arraysHuawei.Arrays[i].Site, arraysHuawei.Arrays[i].Type, arraysHuawei.Arrays[i].Client, model, test)
			pools.Pools = append(pools.Pools, arrayPools.Pools...)
		}

	}
	var telia Client
	telia.Name = "Telia"
	telia.P16Total = 0
	telia.P16Free = 0
	telia.P16InternalTotal = 0
	telia.P16InternalFree = 0
	telia.P16InternalSSDTotal = 0
	telia.P16InternalHDDTotal = 0
	telia.P16InternalSSDFree = 0
	telia.P16InternalHDDFree = 0
	telia.P16InternalSSDMinLun = 0
	telia.P16InternalHDDMinLun = 0
	telia.P16ExternalTotal = 0
	telia.P16ExternalFree = 0
	telia.P16ExternalSSDTotal = 0
	telia.P16ExternalHDDTotal = 0
	telia.P16ExternalSSDFree = 0
	telia.P16ExternalHDDFree = 0
	telia.P16ExternalSSDMinLun = 0
	telia.P16ExternalHDDMinLun = 0
	telia.Z141Total = 0
	telia.Z141Free = 0
	telia.Z141InternalTotal = 0
	telia.Z141InternalFree = 0
	telia.Z141InternalSSDTotal = 0
	telia.Z141InternalHDDTotal = 0
	telia.Z141InternalSSDFree = 0
	telia.Z141InternalHDDFree = 0
	telia.Z141InternalSSDMinLun = 0
	telia.Z141InternalHDDMinLun = 0
	telia.Z141ExternalTotal = 0
	telia.Z141ExternalFree = 0
	telia.Z141ExternalSSDTotal = 0
	telia.Z141ExternalHDDTotal = 0
	telia.Z141ExternalSSDFree = 0
	telia.Z141ExternalHDDFree = 0
	telia.Z141ExternalSSDMinLun = 0
	telia.Z141ExternalHDDMinLun = 0
	telia.Total = 0
	telia.TotalFree = 0

	for _, s := range pools.Pools {
		if s.Client == "Telia" {
			// fmt.Println(s.Type)

			telia.Total += s.PoolCapacity
			telia.TotalFree += s.PoolCapacityFree
			if s.Site == "P16" {
				telia.P16Total += s.PoolCapacity
				telia.P16Free += s.PoolCapacityFree
				if strings.Contains(s.Type, "Internal") {
					telia.P16InternalTotal += s.PoolCapacity
					telia.P16InternalFree += s.PoolCapacityFree
					if strings.Contains(s.Type, "SSD") || strings.Contains(s.Type, "MIX") {
						telia.P16InternalSSDTotal += s.PoolCapacity
						telia.P16InternalSSDFree += s.PoolCapacityFree
						telia.P16InternalSSDMinLun += int(s.PoolCapacityFree / 10000000000000)
					} else if strings.Contains(s.Type, "SAS") {
						telia.P16InternalHDDTotal += s.PoolCapacity
						telia.P16InternalHDDFree += s.PoolCapacityFree
						telia.P16InternalHDDMinLun += int(s.PoolCapacityFree / 10000000000000)
					}
				} else if strings.Contains(s.Type, "Shared") {
					telia.P16ExternalTotal += s.PoolCapacity
					telia.P16ExternalFree += s.PoolCapacityFree
					if strings.Contains(s.Type, "SSD") || strings.Contains(s.Type, "MIX") {
						telia.P16ExternalSSDTotal += s.PoolCapacity
						telia.P16ExternalSSDFree += s.PoolCapacityFree
						telia.P16ExternalSSDMinLun += int(s.PoolCapacityFree / 10000000000000)
					} else if strings.Contains(s.Type, "SAS") {
						telia.P16ExternalHDDTotal += s.PoolCapacity
						telia.P16ExternalHDDFree += s.PoolCapacityFree
						telia.P16ExternalHDDMinLun += int(s.PoolCapacityFree / 10000000000000)
					}
				}
			} else if s.Site == "Z141" {
				telia.Z141Total += s.PoolCapacity
				telia.Z141Free += s.PoolCapacityFree
				if strings.Contains(s.Type, "Internal") {
					telia.Z141InternalTotal += s.PoolCapacity
					telia.Z141InternalFree += s.PoolCapacityFree
					if strings.Contains(s.Type, "SSD") || strings.Contains(s.Type, "MIX") {
						telia.Z141InternalSSDTotal += s.PoolCapacity
						telia.Z141InternalSSDFree += s.PoolCapacityFree
						telia.Z141InternalSSDMinLun += int(s.PoolCapacityFree / 10000000000000)
					} else if strings.Contains(s.Type, "SAS") {
						telia.Z141InternalHDDTotal += s.PoolCapacity
						telia.Z141InternalHDDFree += s.PoolCapacityFree
						telia.Z141InternalHDDMinLun += int(s.PoolCapacityFree / 10000000000000)
					}
				} else if strings.Contains(s.Type, "Shared") {
					telia.Z141ExternalTotal += s.PoolCapacity
					telia.Z141ExternalFree += s.PoolCapacityFree
					if strings.Contains(s.Type, "SSD") || strings.Contains(s.Type, "MIX") {
						telia.Z141ExternalSSDTotal += s.PoolCapacity
						telia.Z141ExternalSSDFree += s.PoolCapacityFree
						telia.Z141ExternalSSDMinLun += int(s.PoolCapacityFree / 10000000000000)
					} else if strings.Contains(s.Type, "SAS") {
						telia.Z141ExternalHDDTotal += s.PoolCapacity
						telia.Z141ExternalHDDFree += s.PoolCapacityFree
						telia.Z141ExternalHDDMinLun += int(s.PoolCapacityFree / 10000000000000)
					}
				}
			} else if s.Site == "Stretched" {

				if strings.Contains(s.PoolName, "P16") {
					telia.P16Total += s.PoolCapacity
					telia.P16Free += s.PoolCapacityFree
					telia.StretchedP16Total += s.PoolCapacity
					telia.StretchedP16Free += s.PoolCapacityFree
					telia.StretchedP16MinLun += int(s.PoolCapacityFree / 10000000000000)

				} else if strings.Contains(s.PoolName, "Z141") {
					telia.Z141Total += s.PoolCapacity
					telia.Z141Free += s.PoolCapacityFree
					telia.StretchedZ141Total += s.PoolCapacity
					telia.StretchedZ141Free += s.PoolCapacityFree
					telia.StretchedZ141MinLun += int(s.PoolCapacityFree / 10000000000000)
				}

			}

		}
	}
	url := "http://xxx/write?db=capacity_metrics"
	ts := fmt.Sprint(time.Now().UnixNano())
	{

		for _, pool := range pools.Pools {
			totalString := "testData,ID=\"" + pool.Id + pool.ArrayName + ",site=" + pool.Site + ",type=" + pool.Type + " Array=\"" + pool.ArrayName + "\",Firmware=\"" + strings.ReplaceAll(strings.ReplaceAll(pool.Firmware, " ", ""), ",", "") + "\",Pool=\"" + pool.PoolName + "\",TotalCapacity=" + fmt.Sprintf("%f", pool.PoolCapacity) + ",FreeCapacity=" + fmt.Sprintf("%f", pool.PoolCapacityFree) + ",UsedCapacity=" + fmt.Sprintf("%f", pool.PoolCapacityUsed) + ",AllocationPCT=" + fmt.Sprintf("%f", pool.PoolCapacityPCT) + " " + ts
			resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewBuffer([]byte(totalString)))
			if err != nil {
				log.Fatalln(err)
			}
			defer resp.Body.Close()
			// // fmt.Println(totalString)
			// var res map[string]interface{}
			// json.NewDecoder(resp.Body).Decode(&res)
			// fmt.Println(fmt.Sprint(resp.StatusCode))
		}
		clientString := "clientData,client=\"" + telia.Name + "\" Total=" + fmt.Sprintf("%f", telia.Total) + ",TotalFree=" + fmt.Sprintf("%f", telia.TotalFree) + ",P16Total=" + fmt.Sprintf("%f", telia.P16Total) + ",P16Free=" + fmt.Sprintf("%f", telia.P16Free) + ",Z141Total=" + fmt.Sprintf("%f", telia.Z141Total) + ",Z141Free=" + fmt.Sprintf("%f", telia.Z141Free) + ",P16InternalTotal=" + fmt.Sprintf("%f", telia.P16InternalTotal) + ",P16InternalFree=" + fmt.Sprintf("%f", telia.P16InternalFree) + ",P16InternalSSDTotal=" + fmt.Sprintf("%f", telia.P16InternalSSDTotal) + ",P16InternalHDDTotal=" + fmt.Sprintf("%f", telia.P16InternalHDDTotal) + ",P16InternalSSDFree=" + fmt.Sprintf("%f", telia.P16InternalSSDFree) + ",P16InternalHDDFree=" + fmt.Sprintf("%f", telia.P16InternalHDDFree) + ",P16InternalSSDMinLun=" + fmt.Sprint(telia.P16InternalSSDMinLun) + ",P16InternalHDDMinLun=" + fmt.Sprint(telia.P16InternalHDDMinLun) + ",P16ExternalTotal=" + fmt.Sprintf("%f", telia.P16ExternalTotal) + ",P16ExternalFree=" + fmt.Sprintf("%f", telia.P16ExternalFree) + ",P16ExternalSSDTotal=" + fmt.Sprintf("%f", telia.P16ExternalSSDTotal) + ",P16ExternalHDDTotal=" + fmt.Sprintf("%f", telia.P16ExternalHDDTotal) + ",P16ExternalSSDFree=" + fmt.Sprintf("%f", telia.P16ExternalSSDFree) + ",P16ExternalHDDFree=" + fmt.Sprintf("%f", telia.P16ExternalHDDFree) + ",P16ExternalSSDMinLun=" + fmt.Sprint(telia.P16ExternalSSDMinLun) + ",P16ExternalHDDMinLun=" + fmt.Sprint(telia.P16ExternalHDDMinLun) + ",Z141InternalTotal=" + fmt.Sprintf("%f", telia.Z141InternalTotal) + ",Z141InternalFree=" + fmt.Sprintf("%f", telia.Z141InternalFree) + ",Z141InternalSSDTotal=" + fmt.Sprintf("%f", telia.Z141InternalSSDTotal) + ",Z141InternalHDDTotal=" + fmt.Sprintf("%f", telia.Z141InternalHDDTotal) + ",Z141InternalSSDFree=" + fmt.Sprintf("%f", telia.Z141InternalSSDFree) + ",Z141InternalHDDFree=" + fmt.Sprintf("%f", telia.Z141InternalHDDFree) + ",Z141InternalSSDMinLun=" + fmt.Sprint(telia.Z141InternalSSDMinLun) + ",Z141InternalHDDMinLun=" + fmt.Sprint(telia.Z141InternalHDDMinLun) + ",Z141ExternalTotal=" + fmt.Sprintf("%f", telia.Z141ExternalTotal) + ",Z141ExternalFree=" + fmt.Sprintf("%f", telia.Z141ExternalFree) + ",Z141ExternalSSDTotal=" + fmt.Sprintf("%f", telia.Z141ExternalSSDTotal) + ",Z141ExternalHDDTotal=" + fmt.Sprintf("%f", telia.Z141ExternalHDDTotal) + ",Z141ExternalSSDFree=" + fmt.Sprintf("%f", telia.Z141ExternalSSDFree) + ",Z141ExternalHDDFree=" + fmt.Sprintf("%f", telia.Z141ExternalHDDFree) + ",Z141ExternalSSDMinLun=" + fmt.Sprint(telia.Z141ExternalSSDMinLun) + ",Z141ExternalHDDMinLun=" + fmt.Sprint(telia.Z141ExternalHDDMinLun) + ",StretchedP16Total=" + fmt.Sprint(telia.StretchedP16Total) + ",StretchedP16Free=" + fmt.Sprint(telia.StretchedP16Free) + ",StretchedP16MinLun=" + fmt.Sprint(telia.StretchedP16MinLun) + ",StretchedZ141Total=" + fmt.Sprint(telia.StretchedZ141Total) + ",StretchedZ141Free=" + fmt.Sprint(telia.StretchedZ141Free) + ",StretchedZ141MinLun=" + fmt.Sprint(telia.StretchedZ141MinLun) + " " + ts
		resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewBuffer([]byte(clientString)))
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()
		// // fmt.Println(clientString)
		// var res map[string]interface{}
		// json.NewDecoder(resp.Body).Decode(&res)
		// fmt.Println(fmt.Sprint(resp.StatusCode))
	}
	logError("Finish")
}
