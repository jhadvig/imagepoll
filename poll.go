package main

import (
	"net/http"
	"log"
	"github.com/fsouza/go-dockerclient"
	"github.com/golang/glog"
	"io/ioutil"
	"net/url"
	"fmt"
	"strings"
	// "encoding/json"

	"flag"
)

const (
	usage = "usage: poller -registry=name -repository=name -image=name -user=name -pw=password"
	indexScheme = "https"
	indexHost = "index.docker.io"
)


var registryName = flag.String("registry", "", "set registry name")
var repositoryName = flag.String("repository", "", "set repository name")
var imageName = flag.String("image", "", "set image name")
var containerID = flag.String("containerID", "", "set image name")
var userName = flag.String("user", "", "set username for the official docker registry")
var password = flag.String("pw", "", "set user password for the official docker registry")

type Tags struct {
	Latest string `json:"latest"`
}

func poll(pollURL *url.URL, token string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", pollURL.String(), nil)
	if token != "" {
		t := fmt.Sprintf("Token %s",token)
		req.Header.Set("Authorization", t)
	} 
	resp, err := client.Do(req)
	defer resp.Body.Close()
	checkErr(err)
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	return err
}

func getTokenAndEndpoint(userName, password, image string) (endpoint, token string) {
	client := &http.Client{}
	url := &url.URL{
		Scheme: indexScheme,
		Host: indexHost,
		Path: fmt.Sprintf("/v1/repositories/%s/images", image),
	}
	req, err := http.NewRequest("GET", url.String(), nil)
	checkErr(err)
	if userName != "" && password != "" {
		req.SetBasicAuth(userName, password)
	}
	req.Header.Set("X-Docker-Token", "true")
	resp, err := client.Do(req)
	checkErr(err)

	endpoint = resp.Header.Get("X-Docker-Endpoints")
	token = resp.Header.Get("X-Docker-Token")
	return
}

func main() {
	flag.Parse()

	if *repositoryName == "" || *imageName == "" {
		log.Fatal(usage)
	}
	image := strings.Join([]string{*repositoryName, *imageName}, "/")
	endpoint := "unix:///var/run/docker.sock"
	dockerClient, err := docker.NewClient(endpoint)
	checkErr(err)

	containerMeta, err := dockerClient.InspectContainer(*containerID)
	checkErr(err)
	imageID := containerMeta.Image
	imageName := containerMeta.Config.Image

	fmt.Printf("\nimageID: %s\n", imageID)
	fmt.Printf("\nimagename: %s\n", imageName)	

	scheme := "http"
	registryEndpoint := *registryName
	registryPath := fmt.Sprintf("/v1/repositories/%s/tags/latest", image)
	token := ""

	if registryEndpoint == "" {
		registryEndpoint, token = getTokenAndEndpoint(*userName, *password, image)
		scheme = "https"
	}

	pollURL := &url.URL{
		Scheme: scheme,
		Host: registryEndpoint,
		Path: registryPath,
	}

	err = poll(pollURL, token)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		glog.Fatalf("%v", err)
	}
}
