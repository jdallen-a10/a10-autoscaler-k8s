# a10-autoscaler-k8s
Autoscale a K8s Deployment based on the Throughput of a Virtual Server on a Thunder ADC.

If you are using a Thunder ADC as a Cloud Application Endpoint-of-Control, all your application traffic will be going through a Thunder ADC SLB Virtual Server. Since all that Cloud Application traffic is running through that SLB, we can us the Traffic Throughput rate to scale up or down the number of Cloud Application Pods that need to be running to suppor that traffic load.

This Proof-of-Concept Thunder Cloud Agent (TCA) implments a simple way to scale a Cloud Application Deployment. Its designed to be used in conjunction with the A10 Thunder Kubernetes Connector to provide for a completely hands-free operation of a Cloud Application.

Documentation on how to config and run the Container is in a PDF file located in this Repo.
