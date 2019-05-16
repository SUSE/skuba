/**
 * This pipeline verifies basic caaspctl deployment, bootstrapping, and adding nodes to a cluster on GitHub Pr
 */

void setBuildStatus(String context, String description, String state) {
    def body = "{\"state\": \"${state}\", " +
               "\"target_url\": \"${BUILD_URL}/display/redirect\", " +
               "\"description\": \"${description}\", " +
               "\"context\": \"${context}\"}"
    def headers = '-H "Content-Type: application/json" -H "Accept: application/vnd.github.v3+json"'
    def url = "https://${GITHUB_TOKEN}@api.github.com/repos/SUSE/caaspctl/statuses/${GIT_COMMIT}"

    sh(script: "curl -X POST ${headers} ${url} -d '${body}'", label: "Sending commit status")
}

pipeline {
    agent { node { label 'caasp-team-private' } }

    environment {
        OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        PLATFORM = 'openstack'
    }

    stages {
        stage('Setting GitHub in-progress status') { steps {
            setBuildStatus('jenkins/caaspctl-integration', 'in-progress', 'pending')
        } }

        stage('Git Clone') { steps {
            deleteDir()
            checkout([$class: 'GitSCM',
                      branches: [[name: "*/${BRANCH_NAME}"], [name: '*/master']],
                      doGenerateSubmoduleConfigurations: false,
                      extensions: [[$class: 'LocalBranch'],
                                   [$class: 'WipeWorkspace'],
                                   [$class: 'RelativeTargetDirectory', relativeTargetDir: 'caaspctl']],
                      submoduleCfg: [],
                      userRemoteConfigs: [[refspec: '+refs/pull/*/head:refs/remotes/origin/PR-*',
                                           credentialsId: 'github-token',
                                           url: 'https://github.com/SUSE/caaspctl']]])

            dir("${WORKSPACE}/caaspctl") {
                sh(script: "git checkout ${BRANCH_NAME}", label: "Checkout PR Branch")
            }
        }}

        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f caaspctl/ci/Makefile pre_deployment', label: 'Pre Deployment')
            sh(script: 'make -f caaspctl/ci/Makefile pr_checks', label: 'PR Checks')
        } }

        stage('Cluster Deployment') { steps {
            sh(script: 'make -f caaspctl/ci/Makefile deploy', label: 'Deploy')
            archiveArtifacts("caaspctl/ci/infra/${PLATFORM}/terraform.tfstate")
            archiveArtifacts("caaspctl/ci/infra/${PLATFORM}/terraform.tfvars")
        } }

        stage('Bootstrap Cluster') { steps {
            sh(script: 'make -f caaspctl/ci/Makefile bootstrap', label: 'Bootstrap')
        } }

        stage('Add Nodes to Cluster') { steps {
            sh(script: 'make -f caaspctl/ci/Makefile add_nodes', label: 'Add Nodes')
        } }
    }
    post {
        always {
            sh(script: 'make --keep-going -f caaspctl/ci/Makefile post_run', label: 'Post Run')
        }
        cleanup {
            dir("${WORKSPACE}") {
                deleteDir()
            }
        }
        failure {
            setBuildStatus('jenkins/caaspctl-integration', 'failed', 'failure')
        }
        success {
            setBuildStatus('jenkins/caaspctl-integration', 'success', 'success')
        }
    }
}
