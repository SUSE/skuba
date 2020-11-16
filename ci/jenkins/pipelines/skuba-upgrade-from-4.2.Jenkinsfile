def cleanup_cluster=true

pipeline {

    agent { label "${WORKER_LABEL}" }
    
    environment {
        SKUBA_BINPATH = "/home/jenkins/go/bin/skuba"
        OPENSTACK_OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        VMWARE_ENV_FILE = credentials('vmware-env')
        TERRAFORM_STACK_NAME = "${BUILD_NUMBER}-${JOB_NAME.replaceAll("/","-")}".take(70)
     }

     stages {

        stage("Build skuba"){
            steps{
                sh(script: "git checkout ${INITIAL_VERSION}")
                sh(script: "make -f Makefile install", label: "Build Skuba ${INITIAL_VERSION}")
            }
        }

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

        stage('upgrade admin node') {
            environment {
                REG_CODE = credentials('scc-regcode-caasp')
            }
            steps {
                echo 'upgrading sles'
                sh "sudo sudo zypper in -y --no-recommends zypper-migration-plugin"
                sh "sudo SUSEConnect -r ${REG_CODE}"
                sh "sudo zypper migration --migration 2 --no-recommends --non-interactive --auto-agree-with-licenses"
                echo "upgrade completed"
            }
        }
    
        stage('reboot'){
            agent { label 'caasp-team-private-integration' } 
            environment {
                JENKINS_ID = credentials('jenkins-ssh')
            }
            steps{
                echo 'rebooting ${WORKER_HOST} ...'
                sh "ssh -i ${JENKINS_ID} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -T jenkins@${WORKER_HOST} sudo reboot &"
               
                sleep 120
 
                timeout(180){
                    waitUntil { script {
                        sh(script: "ssh -i ${JENKINS_ID} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -T jenkins@${WORKER_HOST} sudo systemctl is-active --quiet swarm-client", returnStatus: true) == 0
                    }}
                }
            }
        }

        stage('Inhibit kured reboots') { steps {
            sh(script: 'make -f ci/Makefile inhibit_kured')
            }
        }
        stage('upgrade skuba') { 
            steps {
                sh(script: "sudo zypper in -y --no-recommends go1.13")
                sh(script: "git checkout -f ${TARGET_VERSION}")
                sh(script: "make install", label: "Build ${TARGET_VERSION}")
            }   
        }

        stage('upgrade cluster'){
            environment {
                REG_CODE = credentials('scc-regcode-caasp')
                
            }
            steps {
                sh(script: "make -f ci/Makefile test SUITE='test_upgrade_from_4_2' SKIP_SETUP='deployed'", label: "Test upgrade")
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
            if (Boolean.parseBoolean(env.RETAIN_CLUSTER)) {
                def retention_period= env.RETENTION_PERIOD?env.RETENTION_PERIOD:24
                try{
                    timeout(time: retention_period, unit: 'HOURS'){
                        input(message: 'Waiting '+retention_period +' hours before cleaning up cluster \n. ' +
                                       'Press <abort> to cleanup inmediately, <keep> for keeping it',
                              ok: 'keep')
                        cleanup_cluster = false
                    }
                }catch (err){
                    // either timeout occurred or <abort> was selected
                    cleanup_cluster = true
                }
            }
        }}
        cleanup{ script{
            if(cleanup_cluster){
                sh(script: "make --keep-going -f ci/Makefile cleanup", label: 'Cleanup')
            }
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
        }}
    }
}

