/**
 * This pipeline verifies basic caaspctl deployment, bootstrapping, and adding nodes to a cluster on GitHub Pr
 */

void setBuildStatus(String context, String message, String state) {
  step([
      $class: "GitHubCommitStatusSetter",
      reposSource: [$class: "ManuallyEnteredRepositorySource", url: "https://github.com/SUSE/caaspctl.git"],
      contextSource: [$class: "ManuallyEnteredCommitContextSource", context: context],
      errorHandlers: [[$class: "ChangingBuildStatusErrorHandler", result: "UNSTABLE"]],
      statusResultSource: [ $class: "ConditionalStatusResultSource", results: [[$class: "AnyBuildResult", message: message, state: state]] ]
  ]);
}

pipeline {
    agent { node { label 'caasp-team-private' } }

    environment {
        OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        PARAMS = "stack-type=openstack-terraform no-collab-check"
    }

    stages {
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
        }}

        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f caaspctl/ci/Makefile pre_deployment', label: 'Pre Deployment')
            sh(script: 'make -f caaspctl/ci/Makefile pr_checks', label: 'PR Checks')
        } }

        stage('Cluster Deployment') { steps {
            sh(script: 'make -f caaspctl/ci/Makefile deploy', label: 'Deploy')
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
            sh(script: 'make -f caaspctl/ci/Makefile post_run', label: 'Post Run')
        }
        cleanup {
            dir("${WORKSPACE}") {
                deleteDir()
            }
        }
        failure {
            setBuildStatus("jenkins/caaspctl-integration", "failed", "FAILURE")
        }
        success {
            setBuildStatus("jenkins/caaspctl-integration", "success", "SUCCESS")
        }
    }
}
