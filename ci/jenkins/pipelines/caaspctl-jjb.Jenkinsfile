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
            sh(script: "make -f ci/Makefile stage=info ${PARAMS}", label: 'Info')
        } }
        stage('Setup Environment') { steps {
            sh(script: 'python3 -m venv venv', label: 'Setup Python Virtualenv')
            sh(script: 'venv/bin/pip install jenkins-job-builder==2.10.0', label: 'Install Dependencies')
        } }
        stage('Test Jobs') { steps {
            sh(script: """
                   source ${WORKSPACE}/venv/bin/activate
                   make -f ci/Makefile test_jenkins_jobs
                """, label: 'Test Jenkins Jobs')
            zip archive: true, dir: 'jobs', zipFile: 'jenkins_jobs.zip'
        } }
        stage('Update Jobs') { steps {
            sh(script: """
                   source ${WORKSPACE}/venv/bin/activate
                   make -f ci/Makefile update_jenkins_jobs
                """, label: 'Update Jenkins Jobs')
        } }
    }
    post {
        always {
            cleanWs()
        }
    }
}
