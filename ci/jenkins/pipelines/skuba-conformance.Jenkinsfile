// type of worker required by the job 
def worker_type = 'integration'

node('caasp-team-private-integration') {
    stage('select worker') {
        if (env.BRANCH != 'master') {
            if (env.BRANCH.startsWith('experimental') || env.BRANCH.startsWith('maintenance')) {
                worker_type = env.BRANCH
            }
        }
    }
}

pipeline {
   agent { node { label "caasp-team-private-${worker_type} && e2e" } }

    environment {
        SKUBA_BINPATH = "/home/jenkins/go/bin/skuba"
        OPENSTACK_OPENRC = credentials('ecp-openrc')
        TERRAFORM_STACK_NAME = "${BUILD_NUMBER}-${JOB_NAME.replaceAll("/","-")}".take(70)
        GITHUB_TOKEN = credentials('github-token')
        VMWARE_ENV_FILE = credentials('vmware-env')
        SONOBUOY_VERSION = "v0.17"
    }

    stages {

        stage('Git Clone') { steps {
            deleteDir()
            checkout([$class: 'GitSCM',
                      branches: [[name: "*/${BRANCH}"]],
                      doGenerateSubmoduleConfigurations: false,
                      extensions: [[$class: 'LocalBranch'],
                                   [$class: 'WipeWorkspace'],
                                   [$class: 'RelativeTargetDirectory', relativeTargetDir: 'skuba'],
                                   [$class: 'ChangelogToBranch', options: [compareRemote: "origin", compareTarget: "master"]]],
                      submoduleCfg: [],
                      userRemoteConfigs: [[refspec: '+refs/pull/*/head:refs/remotes/origin/PR-*',
                                           credentialsId: 'github-token',
                                           url: 'https://github.com/SUSE/skuba']]])
        }}
        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f skuba/ci/Makefile pre_deployment', label: 'Pre Deployment')
            sh(script: 'cd skuba; make install; cd ../', label: 'Install skuba')
        } }

        stage('Provision  cluster') { steps {
            sh(script: 'make -f skuba/ci/Makefile provision', label: 'Provision')
        } }

        stage('Deploy cluster') { steps {
            sh(script: 'make -f skuba/ci/Makefile deploy', label: 'Deploy')
            sh(script: 'make -f skuba/ci/Makefile check_cluster', label: 'Check cluster')
        } }

        stage('Inhibit kured reboots') { steps {
            sh(script: 'make -f skuba/ci/Makefile inhibit_kured')
            }
        }

        stage('Conformance Tests') {
            options { timeout(time: 200, unit: 'MINUTES', activity: false) }
            steps {
                sh(script: "skuba/ci/tasks/sonobuoy_e2e.py run --kubeconfig ${WORKSPACE}/test-cluster/admin.conf --sonobuoy-version ${SONOBUOY_VERSION} --mode=certified-conformance", label: 'Run Conformance')
                sh(script: "skuba/ci/tasks/sonobuoy_e2e.py collect --kubeconfig ${WORKSPACE}/test-cluster/admin.conf --sonobuoy-version ${SONOBUOY_VERSION}", label: 'Collect Results')
                sh(script: "skuba/ci/tasks/sonobuoy_e2e.py cleanup --kubeconfig ${WORKSPACE}/test-cluster/admin.conf --sonobuoy-version ${SONOBUOY_VERSION}", label: 'Cleanup Cluster')
            }
        }

    }
    post {
        always {
            archiveArtifacts(artifacts: "skuba/ci/infra/${PLATFORM}/terraform.tfstate", allowEmptyArchive: true)
            archiveArtifacts(artifacts: "skuba/ci/infra/${PLATFORM}/terraform.tfvars.json", allowEmptyArchive: true)
            archiveArtifacts(artifacts: 'testrunner.log', allowEmptyArchive: true)
            archiveArtifacts('results/**/*')
            sh(script: 'make --keep-going -f skuba/ci/Makefile post_run', label: 'Post Run', returnStatus: true)
            zip(archive: true, dir: 'platform_logs', zipFile: 'platform_logs.zip')
            junit('results/plugins/e2e/results/**/*.xml')
        }
        cleanup {
            dir("${WORKSPACE}") {
                deleteDir()
            }
            sh(script: "rm -f ${SKUBA_BINPATH}; ", label: 'Remove built skuba')
        }
    }
}
