/**
 * This pipeline perform basic checks on Pull-requests. (go vet) etc
 */

pipeline {
    agent { node { label 'caasp-team-private' } }

    environment {
        OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        PARAMS = "stack-type=openstack-terraform no-collab-check"
    }

    stages {
        stage('Running go vet') { steps {
            sh("make vet")
        } }

        // TODO: Add here golint later on

    }
    post {
        always {
            cleanWs()
        }
    }
}
