/**
 * This pipeline runs any e2e test parametrized
 */

// type of worker required by the job 
def worker_type = 'integration'

// Repo and registry branch
def branch_repo = ""
def branch_registry = ""

// original (non-branched) registry, only needed when branch_registry is in use
def original_registry = ""

// worker selection labels 
def labels = 'e2e'

node('caasp-team-private-integration') {
    stage('select worker') {
        if (env.BRANCH != 'master') {
            if (env.BRANCH.startsWith('experimental') || env.BRANCH.startsWith('maintenance')) {
                worker_type = env.BRANCH
            }
        }

        // Overrride the worker type if explicitly requested
        if (env.WORKER_TYPE) {
            worker_type = env.WORKER_TYPE  
        }

        // Set additional labels for worker selection
        if (env.WORKER_LABELS) {
            labels = env.WORKER_LABELS
        }

        if (env.REPO_BRANCH){
               branch_repo = "http://download.suse.de/ibs/Devel:/CaaSP:/5:/Branches:/${env.REPO_BRANCH}/SLE_15_SP2"
               branch_registry = "registry.suse.de/devel/caasp/5/branches/${env.REPO_BRANCH}/containers"
               original_registry = "registry.suse.de/devel/caasp/5/containers/containers"
         }
    }
}

pipeline {
   agent { node { label "caasp-team-private-${worker_type} && ${labels}" } }

   environment {
        SKUBA_BINPATH = "/home/jenkins/go/bin/skuba"
        OPENSTACK_OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        VMWARE_ENV_FILE = credentials('vmware-env')
        TERRAFORM_STACK_NAME = "${BUILD_NUMBER}-${JOB_NAME.replaceAll("/","-")}".take(70)
        BRANCH_REPO = "${branch_repo}"
        BRANCH_REGISTRY = "${branch_registry}"
        ORIGINAL_REGISTRY = "${original_registry}"
   }

   stages {
        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: "make -f ci/Makefile pre_deployment", label: 'Pre Deployment')
            sh(script: "make -f Makefile install", label: 'Build Skuba')
        } }

        stage('Provision cluster') {
            steps {
                sh(script: 'make -f ci/Makefile provision', label: 'Provision')
            }
        }

        stage('Deploy cluster') {
            steps {
                sh(script: 'make -f ci/Makefile deploy KUBERNETES_VERSION=${KUBERNETES_VERSION}', label: 'Deploy')
                sh(script: 'make -f ci/Makefile check_cluster', label: 'Check cluster')
            }
        }

        stage('Run Skuba e2e Test') {
            steps {
                sh(script: "make -f ci/Makefile test SUITE=${E2E_MAKE_TARGET_NAME} SKIP_SETUP='deployed'", label: "${E2E_MAKE_TARGET_NAME}")
            }
        }
   }

   post {
        always { script {
            archiveArtifacts(artifacts: "ci/infra/${PLATFORM}/terraform.tfstate", allowEmptyArchive: true)
            archiveArtifacts(artifacts: "ci/infra/${PLATFORM}/terraform.tfvars.json", allowEmptyArchive: true)
            archiveArtifacts(artifacts: 'testrunner.log', allowEmptyArchive: true)
            // only attempt to collect logs if platform was provisioned
            if (fileExists("tfout.json")) {
                archiveArtifacts(artifacts: 'tfout.json', allowEmptyArchive: true)
                sh(script: "make --keep-going -f ci/Makefile gather_logs", label: 'Gather Logs')
                archiveArtifacts(artifacts: 'platform_logs/**/*', allowEmptyArchive: true)
            }
        }}
        cleanup {
            sh(script: "make --keep-going -f ci/Makefile cleanup", label: 'Cleanup')
            dir("${WORKSPACE}@tmp") {
                deleteDir()
            }
            dir("${WORKSPACE}@script") {
                deleteDir()
            }
            dir("${WORKSPACE}@script@tmp") {
                deleteDir()
            }
            dir("${WORKSPACE}") {
                deleteDir()
            }
            sh(script: "rm -f ${SKUBA_BINPATH}; ", label: 'Remove built skuba')
       }
    }
}
