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
            sh(script: "make vet")
        } }

        // TODO: Add here golint later on

    }
    post {
        always {
            cleanWs()
        }
    }
}
