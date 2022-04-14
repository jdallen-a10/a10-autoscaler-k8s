package main

//
//  a10-autoscaler-k8s.go
//   A Proof-of-Concept Thunder Cloud Agent to watch the Throughput of a SLB Virtual Server Port and when it goes
//   over/under a certain threshold (as defined in the config file), adjust the 'replicas' of the defined Kubernetes Deployment.
//   The adjustment algorithm is very simplistic with a simple check to see what the current Throughput rate is, and then
//   compute how many Replicas are required based on the Rate in the configuration file. The program will then do an API call
//   to Kubernetes, if an adjustment is needed.
//
//  John D. Allen
//  Global Solutions Architect - Cloud, IoT, & Automation
//  A10 Networks, Inc.
//
import (
	"flag"
	"fmt"
	"io/ioutil"
	"k8sgo"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"a10/axapi"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Configuration struct {
	Debug    int           `yaml:"debug"`
	Interval time.Duration `yaml:"check_interval"`
	Timeout  time.Duration `yaml:"cmd_timeout"`
	Cluster  struct {
		IP         string `yaml:"ip"`
		Port       int    `yaml:"port"`
		Auth_Token string `yaml:"auth_token"`
		Deployment string `yaml:"deployment"`
		Namespace  string `yaml:"namespace"`
		Min_Pods   int    `yaml:"min_pods"`
		Max_Pods   int    `yaml:"max_pods"`
	} `yaml:"cluster"`
	Thunder struct {
		IP        string `yaml:"ip"`
		Port      int    `yaml:"port"`
		Secret    string `yaml:"secret"`
		Secret_NS string `yaml:"secret_namespace"`
		SLB       string `yaml:"slb"`
		SLB_Port  string `yaml:"slb_port"`
		Rate      uint64 `yaml:"rate"`
	} `yaml:"thunder"`
}

// Global Vars
var DEBUG int
var CFG_FILE string

//---------------------------------------------------------------------------------
// getYamlConfig() - Grab configuration variables from the config YAML file
func getYamlConfig(fn string) (Configuration, error) {
	var c Configuration

	yamlFile, err := ioutil.ReadFile(fn)
	if err != nil {
		return Configuration{}, err
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return Configuration{}, err
	}

	return c, nil
}

//---------------------------------------------------------------------------------
// procLoop()  --  Processing Loop
//  This is the main processing loop used to watch the SLB rates and scale the
//  Deployment as needed.
func procLoop(d axapi.Device, c k8sgo.Cluster, cfg Configuration) {
	//
	// Look up the SLB defined in the configuration and get its current rates
	port, err := d.GetVSThroughput(cfg.Thunder.SLB, cfg.Thunder.SLB_Port)
	if err != nil {
		log.Error(err.Error())
		// Abort here?
	}
	// Look up current number of replicas for the defined Deployment
	y, err := c.GetDeploymentStatus(cfg.Cluster.Deployment, cfg.Cluster.Namespace)
	if err != nil {
		log.Fatal(err.Error())
		ending(2)
	}
	if cfg.Debug > 8 {
		fmt.Println("---------------------------")
		fmt.Println("replicas running = ", y.CurrentReplicas)
		fmt.Println("throughput = ", port.Throughput)
	}

	//
	// Compare SLB rate with defined Rate & compute number of replicas required
	rate := port.Throughput / 1000 // Make it Kbps
	rpl := int(rate / cfg.Thunder.Rate)

	if cfg.Debug > 8 {
		fmt.Println("replicas needed = ", rpl)
	}
	//
	// Adjust the number of Replicas, if needed.
	if rpl != y.CurrentReplicas {
		//
		// Scaling UP:
		// This is very simple, as all we really need to do is tell Kubernetes how many more Pods should
		// be started up and running. The Thunder ADC via the Thunder Kubernetes Connector, will be told
		// of the new Pods and will start sending traffic to them.
		// Scaling DOWN:
		// This is not so simple.  We can always just tell Kubernetes to scale down the number of replicas,
		// but the problem with this is that connections that are "in flight" will be suddenly cut off if they
		// are connected to one of the shut-down Pods! The first version of this PoC code just does that...but
		// what really needs to happen is that this program needs to pick one of the Pods and stop sending
		// connections to it. If Thunder ADC is only connecting at the K8s Worker node level, it will be impossible
		// to actually pick the Pod to stop, as the individual Pods are not configured to Thunder. If using IPinIP
		// tunnels down to the Pod level on K8s, then you CAN pick the individual Pods to stop, tell Thunder ADC to
		// stop sending new traffic to that Pod, wait for it to show zero connections, THEN tell K8s to stop THAT
		// PARTICULAR POD.
		if rpl == 0 { // If no traffic, just set to Min_Pods to avoid repeated warnings.
			rpl = cfg.Cluster.Min_Pods
		}
		if rpl < cfg.Cluster.Min_Pods {
			log.Warn("Tried to adjust Replicas below minimum. Adjusting to Minimum.")
			rpl = cfg.Cluster.Min_Pods
		}
		if rpl > cfg.Cluster.Max_Pods {
			log.Warn("Tried to adjust Replicas above maximum. Adjusting to Maximum.")
			rpl = cfg.Cluster.Max_Pods
		}
		if rpl != y.CurrentReplicas { // Check if we still need to adjust
			// Make the adjustment
			out := "Adjusting Deployment '" + y.Name + "' to " + strconv.Itoa(rpl) + " Replicas."
			log.Info(out)
			y, err = c.AdjustDeployment(y, rpl)
			if err != nil {
				log.Error(err.Error())
				// abort here?
			}
			//
			//  Pause here to check and make sure the Cluster adjusts the number of Replicas for the Deployment correctly
			go func() {
				timeout := time.After(cfg.Timeout * time.Second)
				ticker := time.Tick(500 * time.Millisecond)
				for {
					select {
					case <-timeout:
						log.Error("Adjustment of Replicas Timed Out")
						return
					case <-ticker: // Check every half second
						y, err := c.GetDeploymentStatus(y.Name, y.Namespace)
						if err != nil {
							log.Fatal(err.Error())
							return
						}
						if y.CurrentReplicas == rpl { // CurrentReplicas is equal to computed number or Replicas required.
							log.Info("Adjustment of Cluster is Finished")
							return
						}
					}
				}
			}() // This go func() allows the time.Tick() channel to close on the return, and stop firing.
		}

	}
}

//---------------------------------------------------------------------------------
func main() {
	//
	// Setup the logging with Timestamps
	customFormat := new(log.TextFormatter)
	customFormat.TimestampFormat = "2006-01-02 15:04:05" // Yes, it MUST be THIS string!
	customFormat.FullTimestamp = true
	log.SetFormatter(customFormat)
	log.Info("A10 Kubernetes Autoscaler Starting...")

	//
	// Handle Interrrupts
	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigchan
		ending(1)
	}()

	//
	// Process commandline args
	x1 := flag.Int("debug", 0, "Debugging Level")
	x2 := flag.String("config", "./config.yaml", "Configuration File Path")
	flag.Parse()
	DEBUG = *x1
	CFG_FILE = *x2

	// Parse config file
	config, err := getYamlConfig(CFG_FILE)
	if err != nil {
		log.Fatal(err)
	}
	//
	//  command line overrides config file setting for Debug Level.
	if DEBUG != 0 {
		config.Debug = DEBUG
	}

	// Query K8s Cluster
	c := k8sgo.Cluster{}
	c.URL = config.Cluster.IP + ":" + strconv.Itoa(config.Cluster.Port)
	c.Token = config.Cluster.Auth_Token
	// Make sure we can talk to it
	_, err = c.GetAllPods()
	if err != nil {
		log.Fatal(err.Error())
		ending(2)
	} else {
		log.Info("Connected to Kubernetes Cluster")
	}

	//
	// Get K8s Secret with Thunder credentials
	secret, err := c.GetSecret(config.Thunder.Secret, config.Thunder.Secret_NS)
	if err != nil {
		log.Fatal(err.Error())
	}
	// fmt.Println(secret.User)
	// fmt.Println(secret.Passwd)

	//
	// Connect to Thunder node
	d := axapi.Device{}
	ap := config.Thunder.IP + ":" + strconv.Itoa(config.Thunder.Port)
	d.Address = ap
	d.Username = secret.User
	d.Password = secret.Passwd
	d, err = d.Login()
	if err != nil {
		log.Fatal(err.Error())
	} else {
		log.Info("Connected to Thunder Device")
	}
	defer d.Logoff()

	//
	//  Loop for continous rate checking
	procLoop(d, c, config) // First time through, call direct to avoid initial delay
	RunProcLoop(d, c, config)

}

//---------------------------------------------------------------------------------
//  RunProcLoop() - Handles the timing of calling the Processing Loop
func RunProcLoop(d axapi.Device, c k8sgo.Cluster, cfg Configuration) {
	// Run forever.....
	interval := time.Second * cfg.Interval
	for range time.Tick(interval) {
		procLoop(d, c, cfg)
	}
}

//---------------------------------------------------------------------------------
// ending()  --  Tidy up any loose ends before exiting the program.
func ending(sig int) {
	// Function to shut things down and end program.
	os.Exit(sig)
}
