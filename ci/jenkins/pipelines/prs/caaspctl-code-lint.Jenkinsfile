/**
 * This pipeline perform basic checks on Pull-requests. (go vet) etc
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
            checkout([$class: 'GitSCM',
                      branches: [[name: "*/pr/${CHANGE_ID}"], [name: '*/master']],
                      doGenerateSubmoduleConfigurations: false,
                      extensions: [[$class: 'LocalBranch'],
                                   [$class: 'WipeWorkspace'],
                                   [$class: 'PreBuildMerge', options: [mergeRemote: 'origin', mergeTarget: 'master']],
                                   [$class: 'RelativeTargetDirectory', relativeTargetDir: 'caaspctl']],
                      submoduleCfg: [],
                      userRemoteConfigs: [[refspec: '+refs/pull/*/head:refs/remotes/origin/pr/*',
                                           credentialsId: 'github-token',
                                           url: 'https://github.com/SUSE/caaspctl']]])
        } }

        stage('Running go vet') { steps {
            sh("make vet")
        } }

        // TODO: Add here golint later on

    }
    post {
        always {
            cleanWs()
        }
        failure {
            setBuildStatus("jenkins/caaspctl-code-lint", "failed", "FAILURE")
        }
        success {
            setBuildStatus("jenkins/caaspctl-code-lint", "success", "SUCCESS")
        }
    }
}
