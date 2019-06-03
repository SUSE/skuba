
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
username: "" # User deployed stack name
openrc: "" # Path to openrc.sh file
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
  -l, --log             gather logs from nodes
  -v YAML_PATH, --vars YAML_PATH
                        path for platform yaml file. Default is
                        vars/openstack.yaml in
                        {workspace}/ci/infra/testrunner. eg) -v
                        vars/myconfig.yaml
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
3. Use ginkgo for cluster creation.
