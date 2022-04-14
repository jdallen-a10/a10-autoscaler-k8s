package k8sgo

import (
	"crypto/tls"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

//
//
//

type Cluster struct {
	URL   string
	Token string
}

// _restCall is the basic API callout function
//-----------------------------------------------------------------------------
func _restCall(c Cluster, url string, method string, payload *strings.Reader) ([]byte, error) {
	var body []byte

	u := "https://" + c.URL + url
	if method == "" {
		method = "GET"
	}
	if payload == nil {
		payload = strings.NewReader("")
	}

	// Skip insecure SSL verify returns -- for now
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	// set the HTTPS request
	req, err := http.NewRequest(method, u, payload)
	if err != nil {
		return []byte{}, err
	}

	if strings.HasSuffix(url, "/scale") {
		req.Header.Add("Content-Type", "application/strategic-merge-patch+json")
	} else {
		req.Header.Add("Content-Type", "application/json")
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.Token)

	res, err := client.Do(req)

	if err != nil {
		return []byte{}, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	//fmt.Println(res)
	if res.StatusCode > 299 {
		return []byte{}, errors.New(res.Status)
	}

	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

//---------------------------------------------------------------------------------------
func (c Cluster) SetURL(url string) Cluster {
	c.URL = url
	return c
}

//---------------------------------------------------------------------------------------
func (c Cluster) SetToken(token string) Cluster {
	c.Token = token
	return c
}

// getAllPods()  --  Return names of all pods on the Cluster
//---------------------------------------------------------------------------------------
func (c Cluster) GetAllPods() ([]string, error) {
	url := "/api/v1/pods"
	body, err := _restCall(c, url, "GET", nil)
	if err != nil {
		return []string{}, err
	}
	b := []string{}
	for _, v := range gjson.GetBytes(body, "items.#.metadata.name").Array() {
		b = append(b, v.String())
	}
	return b, nil
}

// GetDeploymentNames()  -- Return all Deployment Names
//---------------------------------------------------------------------------------------
func (c Cluster) GetDeploymentNames() ([]string, error) {
	url := "/apis/apps/v1/deployments"
	body, err := _restCall(c, url, "GET", nil)
	if err != nil {
		return []string{}, err
	}
	// b := gjson.GetBytes(body, "items.#.metadata.name").Array()
	// fmt.Println(">>>", b)
	// return b, nil

	b := []string{}
	for _, v := range gjson.GetBytes(body, "items.#.metadata.name").Array() {
		//fmt.Println(v.String())
		b = append(b, v.String())
	}
	return b, nil
}

// GetDeploymentStatus()
//---------------------------------------------------------------------------------------
type Deployment struct {
	Name            string
	Namespace       string
	MinReplicas     int
	CurrentReplicas int
}

func (c Cluster) GetDeploymentStatus(dep string, ns string) (Deployment, error) {
	var d Deployment
	url := "/apis/apps/v1/namespaces/" + ns + "/deployments/" + dep
	body, err := _restCall(c, url, "GET", nil)
	if err != nil {
		return d, err
	}
	d.Name = gjson.GetBytes(body, "metadata.name").Str
	d.Namespace = gjson.GetBytes(body, "metadata.namespace").Str
	d.CurrentReplicas = int(gjson.GetBytes(body, "spec.replicas").Int())
	return d, nil
}

// AdjustDeployment()
//---------------------------------------------------------------------------------------
func (c Cluster) AdjustDeployment(d Deployment, num int) (Deployment, error) {
	//
	//  'num' is the number of Replicas that should be running after the adjustment.
	var n int
	if num == d.CurrentReplicas {
		return d, nil
	}
	if num < d.MinReplicas {
		n = d.MinReplicas
	} else {
		n = num
	}

	// Adjustment needed
	if d.Name == "" {
		return d, errors.New("d.Name cannot be blank in Deployment structure")
	}
	if d.Namespace == "" {
		return d, errors.New("d.Namespace cannot be blank in Deployment structure")
	}
	pl := strings.NewReader("{\n\"spec\":{\n\"replicas\": " + fmt.Sprint(n) + "\n}\n}")
	url := "/apis/apps/v1/namespaces/" + d.Namespace + "/deployments/" + d.Name + "/scale"
	body, err := _restCall(c, url, "PATCH", pl)
	//fmt.Println(string(body))
	if err != nil {
		return d, err
	} else {
		x := int(gjson.GetBytes(body, "spec.replicas").Int())
		if x != num {
			return d, errors.New("Pod Replicas adjustment failed - Required: " + fmt.Sprint(num) + " Active: " + fmt.Sprint(x))
		}
	}

	// Adjustment successful; update Deployment
	d.CurrentReplicas = num
	return d, nil
}

// GetSecret()
//---------------------------------------------------------------------------------------
type Secret struct {
	User   string `json:"username"`
	Passwd string `json:"password"`
}

func (c Cluster) GetSecret(name string, ns string) (Secret, error) {
	var s Secret

	url := "/api/v1/namespaces/" + ns + "/secrets/" + name
	body, err := _restCall(c, url, "GET", nil)
	if err != nil {
		return s, err
	}

	data := []byte(gjson.GetBytes(body, "data").String())
	err = json.Unmarshal(data, &s)
	if err != nil {
		return s, err
	}

	// data is in Base64...decode it before return
	sdec, err := b64.StdEncoding.DecodeString(s.User)
	if err != nil {
		return s, err
	}
	s.User = string(sdec)

	sdec, err = b64.StdEncoding.DecodeString(s.Passwd)
	if err != nil {
		return s, err
	}
	s.Passwd = string(sdec)

	return s, nil
}
