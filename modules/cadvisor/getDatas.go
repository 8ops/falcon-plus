package main

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"github.com/golang/glog"
)

func getCpuNum(dockerdata string) {
	cpuNum = 1
	tmp := getBetween(dockerdata, `"CPU=`, `",`)
	if tmp != "" {
		cpuNum, _ = strconv.ParseInt(tmp, 10, 32)
		if cpuNum == 0 {
			cpuNum = 1
		}
	}
}

func getTag() string {
	//FIXMI:some other message for container
	return ""
}

func getMemLimit(str string) string {
	return getBetween(str, `"memory":{"limit":`, `,"`)
}

func getBetween(str, start, end string) string {
	res := regexp.MustCompile(start + `(.+?)` + end).FindStringSubmatch(str)
	if len(res) <= 1 {
		glog.Error(errors.New("regexp len < 1"), start+" "+end)
		return ""
	}
	return res[1]
}

func getLocalIp() (string, error) {
	conn, err := net.Dial("udp", "ntp.ops.gat:55555")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0], nil
}

func getCadvisorData() (string, error) {
	var (
		resp *http.Response
		err  error
		body []byte
	)
	//url := "http://localhost:" + CadvisorPort + "/api/v1.2/docker"
	ip, err := getLocalIp()
	if err != nil {
		return "", err
	}
	url := "http://" + ip + ":" + CadvisorPort + "/api/v1.2/docker"
	if resp, err = http.Get(url); err != nil {
		glog.Error("Get err in getCadvisorData")
		return "", err
	}
	defer resp.Body.Close()
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		glog.Error("ReadAll err in getCadvisorData")
		return "", err
	}

	return string(body), nil
}

func getUsageData(cadvisorData string) (ausge, busge string) {
	ausge = ""
	busge = ""
	lns := len(strings.Split(cadvisorData, `{"timestamp":`))
	if lns > 1 {
		ausge = strings.Split(cadvisorData, `{"timestamp":`)[1]
		busge = strings.Split(cadvisorData, `{"timestamp":`)[lns-1]
		countNum = lns - 1
	}
	/*
	if len(strings.Split(cadvisorData, `{"timestamp":`)) < 11 {
		glog.Info("b: ",len(strings.Split(cadvisorData, `{"timestamp":`)))
		countNum = 1
		busge = strings.Split(cadvisorData, `{"timestamp":`)[2]
		glog.Info("busge: ",busge)
	} else {
		glog.Info("c: ",len(strings.Split(cadvisorData, `{"timestamp":`)))
		busge = strings.Split(cadvisorData, `{"timestamp":`)[10]
		countNum = 10
		glog.Info("busge: ",busge)
	}
	*/

	return ausge, busge
}

func getContainerId(cadvisorData string) string {

	getContainerId1 := strings.Split(cadvisorData, `],"namespace"`)
	getContainerId2 := strings.Split(getContainerId1[0], `","`)
	getContainerId3 := strings.Split(getContainerId2[1], `"`)
	containerId := getContainerId3[0]

	return containerId
}

func getEndPoint(DockerData string) string {
	// find pause;continue
	pauseName := getBetween(DockerData, `"Entrypoint":\["\/`, `"\],`)
	if pauseName == "pause" {
		return "pause"
	}
	//get endpoint from env first
	endPoint := getBetween(DockerData, `"EndPoint=`, `",`)
	if endPoint != "" {
		return endPoint
	}
	// get endporint from docker hostname
	endPoint = getBetween(DockerData, `"Hostname":"`, `",`)
	if endPoint != "" {
		return endPoint
	}

	filepath := getBetween(DockerData, `"HostsPath":"`, `",`)
	buf := make(map[int]string, 6)
	inputFile, inputError := os.Open(filepath)
	if inputError != nil {
		glog.Error(inputError, "getEndPoint open file err"+filepath)
		return ""
	}
	defer inputFile.Close()

	inputReader := bufio.NewReader(inputFile)
	lineCounter := 0
	for i := 0; i < 2; i++ {
		inputString, readerError := inputReader.ReadString('\n')
		if readerError == io.EOF {
			break
		}
		lineCounter++
		buf[lineCounter] = inputString
	}
	hostname := strings.Split(buf[1], "	")[0]
	hostname = strings.Replace(hostname, "\n", " ", -1)
	return hostname
}

func getDockerData(containerId string) (string, error) {
	str, err := RequestUnixSocket("/containers/"+containerId+"/json", "GET")
	if err != nil {
		glog.Error("getDockerData err")
	}
	return str, nil
}

func RequestUnixSocket(address, method string) (string, error) {
	DOCKER_UNIX_SOCKET := "unix:///var/run/docker.sock"
	// Example: unix:///var/run/docker.sock:/images/json?since=1374067924
	unix_socket_url := DOCKER_UNIX_SOCKET + ":" + address
	u, err := url.Parse(unix_socket_url)
	if err != nil || u.Scheme != "unix" {
		glog.Error("Error to parse unix socket url " + unix_socket_url)
		return "", err
	}

	hostPath := strings.Split(u.Path, ":")
	u.Host = hostPath[0]
	u.Path = hostPath[1]

	conn, err := net.Dial("unix", u.Host)
	if err != nil {
		glog.Error("Error to connect to" + u.Host)
		// fmt.Println("Error to connect to", u.Host, err)
		return "", err
	}

	reader := strings.NewReader("")
	query := ""
	if len(u.RawQuery) > 0 {
		query = "?" + u.RawQuery
	}

	request, err := http.NewRequest(method, u.Path+query, reader)
	if err != nil {
		glog.Error("Error to create http request")
		// fmt.Println("Error to create http request", err)
		return "", err
	}

	client := httputil.NewClientConn(conn, nil)
	response, err := client.Do(request)
	if err != nil {
		glog.Error("Error to achieve http request over unix socket")
		// fmt.Println("Error to achieve http request over unix socket", err)
		return "", err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		glog.Error("Error, get invalid body in answer")
		// fmt.Println("Error, get invalid body in answer")
		return "", err
	}

	defer response.Body.Close()

	return string(body), err
}
