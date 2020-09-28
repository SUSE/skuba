def worker_name
def worker_host
def keep_worker = false

pipeline {
    agent { node {label 'caasp-team-private-integration && e2e' }}

    environment {
        OPENRC = credentials('ecp-openrc')
        JENKINS_ID = credentials('jenkins-ssh')
        CAASP_CI   = credentials('ci.suse.de.caasp-user')
    }
    
    stages {
        stage('cleanup'){
            steps {
                dir("${WORKSPACE}/credentials") {
                    deleteDir()
                }
            }
        }
        stage('Setup') {
            steps {
                // Get automation code from a GitHub repository
                git url: 'https://github.com/SUSE/caasp-automation.git', credentialsId: 'github-token'

                // setup credentials
                dir('jenkins-worker'){
                    sh 'mkdir -p credentials >> /dev/null || true'
                    sh "cp ${env.JENKINS_ID} credentials/jenkins"
                    sh "cp ${env.OPENRC} credentials/ecp.openrc"
                    sh "cp ${env.CAASP_CI} credentials/ci.suse.de.caasp-user"
                }
            }
        }
        
        stage('Create-worker') {
            steps { script {
                worker_name = "caasp-team-${BUILD_NUMBER}-dynamic-worker"
                dir('jenkins-worker'){
                    sh "./manage-workers.sh -c --worker-name ${worker_name} --worker-type dynamic --worker-template ${WORKER_TEMPLATE} --labels ${worker_name} --ssh-key ${env.JENKINS_ID}"
                    worker_host  = sh(script: "./manage-workers.sh -i --worker-name ${worker_name}", returnStdout: true).trim()
                    echo "woker ip ${worker_host}"
                }
            }}
        }
        stage('run-test'){
            steps { script {
                build   job: env.DOWNSTREAM_JOB, 
                        wait: true, 
                        propagate: true,
                        parameters: [
                            string(name: 'WORKER_LABEL', value: "${worker_name}"),
                            string(name: 'WORKER_HOST', value: "${worker_host}"),
                            string(name: 'BRANCH', value: env.BRANCH),
                            string(name: 'RETAIN_CLUSTER', value: env.RETAIN_CLUSTER),
                            string(name: 'RETENTION_PERIOD', value: env.RETENTION_PERIOD),
                            string(name: 'INITIAL_VERSION', value: env.INITIAL_VERSION),
                            string(name: 'TARGET_VERSION', value: env.TARGET_VERSION)
                        ]
            }}    
        }
    }
    post {
        always { script {
            if (Boolean.parseBoolean(env.RETAIN_WORKER)) {
                def retention_period= env.RETENTION_PERIOD?env.RETENTION_PERIOD:24
                try{
                    timeout(time: retention_period, unit: 'HOURS'){
                        input(message: 'Waiting '+retention_period +' hours before cleaning up worker \n. ' +
                                        'Press <abort> to cleanup inmediately, <keep> for keeping it',
                                ok: 'keep')
                        keep_worker = true
                    }
                }catch (err){
                    // either timeout occurred or <abort> was selected
                    keep_worker = false
                }
            }
        }}
        cleanup { script {
            if (!keep_worker){
                dir('jenkins-worker'){
                    sh "./manage-workers.sh -d --worker-name ${worker_name} || true"
                }
            }
            dir("${WORKSPACE}") {
                deleteDir()
            }
            dir("${WORKSPACE}@tmp") {
                deleteDir()
            }
            dir("${WORKSPACE}@script") {
                deleteDir()
            }
        }}
    }
}
