
# Testrunner

### Local Dev Machine Setup For Openstack
1. Create a private directory 
```
sudo mkdir /Private
``` 
2. Download container-openrc.sh from openstack, change the file, and store the file in /Private

3. Add openstack password to downloaded container-openrc.sh like the following, save and move to /Private
```
export OS_USERNAME="YOUR USERNAME"
# With Keystone you pass the keystone password.
#echo "Please enter your OpenStack Password for project $OS_PROJECT_NAME as user $OS_USERNAME: "
#read -sr OS_PASSWORD_INPUT
#export OS_PASSWORD=$OS_PASSWORD_INPUT
export OS_PASSWORD="YOUR PASSWORD"
```
4. Edit and update `ci/infra/testrunner/vars/openstack.yaml`
```
workspace: "" # The top folder where skuba is stored
username: ""  # User deployed stack name
openrc: ""    # Path to openrc.sh file
skuba:        # skuba locations
  srcpath:    # Path to skuba srch project (defaults to `./skuba`)
  binpath     # Path to skuba bin directory (defaults to `<workspace>/go/bin/`)
```

5. Use testrunner
```
./testrunner -h
Starting ./testrunner script
usage: 
    This script is meant to be run manually on test servers, developer desktops, or Jenkins.
    This script supposed to run on python virtualenv from testrunner. Requires root privileges.
    Warning: it removes docker containers, VMs, images, and network configuration.
    
       [-h] [-z] [-i] [-x] [-t] [-c] [-b] [-k] [-a] [-r] [-l] [-v YAML_PATH]
       [-m NUM_MASTER] [-w NUM_WORKER]

optional arguments:
  -h, --help            show this help message and exit
  -z, --git-rebase      git rebase to master
  -i, --info            ip info
  -x, --cleanup         cleanup created skuba environment
  -t, --terraform-apply
                        deploy nodes for cluster in your configured platform
                        e.g) openstack, vmware, ..
  -c, --create-skuba
                        create skuba environment
                        {workspace}/go/src/github.com/SUSE/skuba and build
                        skuba in that directory
  -b, --bootstrap       bootstrap k8s cluster with deployed nodes in your
                        platform
  -k, --status          check K8s cluster status
  -a, --add-nodes       add nodes in k8s cluster. Default values are -m=1,
                        -w=1
  -r, --remove-nodes    remove nodes in k8s cluster. default values are -m=1,
                        -w=1
  -l, --log             gather logs from nodes
  -v YAML_PATH, --vars YAML_PATH
                        path for platform yaml file. Default is
                        vars/openstack.yaml in
                        {workspace}/ci/infra/testrunner. eg) -v
                        vars/myconfig.yaml
  -m NUM_MASTER, --master NUM_MASTER
                        number of masters to add or delete. It is dependening
                        on number of deployed master nodes in your yaml file.
                        Default value is 1. eg) -m 2
  -w NUM_WORKER, --worker NUM_WORKER
                        number of workers to add or delete. It is dependening
                        on number of deployed worker nodes in your yaml file.
                        Default value is 1 eg) -w 2


```


### Jenkins Machine Setup
1. In your Jenkins file, you need to set up environment variables. Then These environment variables will replace
variables in yaml file.

This pipeline script is same as openrc: "/Private/container-openrc.sh" in openstack.yaml.
As default, Jenkins has WORKSPACE environment variable so that workspace will be replaced in Jenkins workspace
```
   environment {
        OPENRC = "/Private/container-openrc.sh "
        GITHUB_TOKEN = readFile("/Private/github-token").trim()
   }
```


### Step to create K8s Cluster and start to use K8s cluster 
1. Cleanup before deploying nodes
```ci/infra/testrunner/testrunner -x ``` 
2. Deploy nodes in openstack  
```ci/infra/testrunner/testrunner -t ```  
3. Create skuba env and Build skuba and store in go bin dir
```ci/infra/testrunner/testrunner -c ```
4. Bootstraping a cluster
```ci/infra/testrunner/testrunner -b ```

Once bootstrapping is done you will be ready to use K8s cluster.

5. To extend the cluster, you can add more node with 
```ci/infra/testrunner/testrunner -a -m 2 -w 2 ```

6. Use K8s
Once your nodes are bootstrapped, {worksapce}/test-cluster folder will be created. Inside test-cluster, Your kubeconfig file will be located in with the name of admin.conf in test-cluster folder.
```
chang@~/Workspace/vNext/test-cluster$ kubectl get pods --all-namespaces --kubeconfig=./admin.conf
NAMESPACE     NAME                                  READY     STATUS    RESTARTS   AGE
kube-system   cilium-6mnrh                          1/1       Running   0          3m
kube-system   cilium-z9rqm                          1/1       Running   0          3m
kube-system   coredns-559fbd6bb4-gw7cn              1/1       Running   0          4m
kube-system   coredns-559fbd6bb4-jqt4r              1/1       Running   0          4m
kube-system   etcd-my-master-0                      1/1       Running   0          3m
kube-system   kube-apiserver-my-master-0            1/1       Running   0          3m
kube-system   kube-controller-manager-my-master-0   1/1       Running   0          3m
kube-system   kube-proxy-782z2                      1/1       Running   0          4m
kube-system   kube-proxy-kf7g5                      1/1       Running   0          3m
kube-system   kube-scheduler-my-master-0            1/1       Running   0          3m
```
