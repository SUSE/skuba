// this pipeline update all jenkins pipelines via jenkins job builder plugin

void setBuildStatus(String context, String description, String state) {
    def body = "{\"state\": \"${state}\", " +
            "\"target_url\": \"${BUILD_URL}/display/redirect\", " +
            "\"description\": \"${description}\", " +
            "\"context\": \"${context}\"}"
    def headers = '-H "Content-Type: application/json" -H "Accept: application/vnd.github.v3+json"'
    def url = "https://${GITHUB_TOKEN}@api.github.com/repos/SUSE/skuba/statuses/${GIT_COMMIT}"

    sh(script: "curl -X POST ${headers} ${url} -d '${body}'", label: "Sending commit status")
}

pipeline {
    agent { node { label 'caasp-team-private' } }
    environment {
        JENKINS_JOB_CONFIG = credentials('jenkins-job-config')
        REQUESTS_CA_BUNDLE = "/var/lib/ca-certificates/ca-bundle.pem"
    }
    stages {
        stage('Setting GitHub in-progress status') { steps {
            setBuildStatus('jenkins/skuba-jjb-validation', 'in-progress', 'pending')
        } }

        stage('Info') { steps {
            sh(script: "make -f skuba/ci/Makefile info", label: 'Info')
        } }

        stage('Setup Environment') { steps {
            sh(script: 'python3 -m venv venv', label: 'Setup Python Virtualenv')
            sh(script: 'venv/bin/pip install jenkins-job-builder==2.10.0', label: 'Install Dependencies')
        } }

        stage('Test Jobs') { steps {
            sh(script: """
                   source ${WORKSPACE}/venv/bin/activate
                   make -f skuba/ci/Makefile test_jenkins_jobs
                """, label: 'Test Jenkins Jobs')
            zip archive: true, dir: 'skuba/ci/jenkins/jobs', zipFile: 'jenkins_jobs.zip'
        } }
    }
    post {
        cleanup {
            dir("${WORKSPACE}") {
                deleteDir()
            }
        }
        failure {
            setBuildStatus('jenkins/skuba-jjb-validation', 'failed', 'failure')
        }
        success {
            setBuildStatus('jenkins/skuba-jjb-validation', 'success', 'success')
        }
    }
}
