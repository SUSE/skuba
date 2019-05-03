// this pipeline update all jenkins pipelines via jenkins job builder plugin
pipeline {
    agent { node { label 'caasp-team-private' } }
    environment {
        JENKINS_JOB_CONFIG = credentials('jenkins-job-config')
        REQUESTS_CA_BUNDLE = "/var/lib/ca-certificates/ca-bundle.pem"
        PARAMS = "openrc=."
    }
    stages {
        stage('Info') { steps {
            sh "caaspctl/ci/infra/testrunner/testrunner stage=info ${PARAMS}"
        } }
        stage('Setup Environment') { steps {
            sh "python3 -m venv venv"
            sh "venv/bin/pip install jenkins-job-builder==2.10.0"
            sh "cp ${JENKINS_JOB_CONFIG} caaspctl/ci/jenkins/jenkins_jobs.ini"
        } }
        stage('Test Jobs') { steps {
            dir('caaspctl/ci/jenkins') {
                sh """
                   source ${WORKSPACE}/venv/bin/activate
                   make test_jobs
                """
                zip archive: true, dir: 'jobs', zipFile: 'jenkins_jobs.zip'
            }
        } }
        stage('Update Jobs') { steps {
            dir('caaspctl/ci/jenkins') {
                sh """
                   source ${WORKSPACE}/venv/bin/activate
                   make update_jobs
                """
            }
        } }
    }
    post {
        always {
            cleanWs()
        }
    }
}
