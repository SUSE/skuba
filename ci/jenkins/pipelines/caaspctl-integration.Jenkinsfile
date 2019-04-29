/**
 * This pipeline verifies basic caaspctl deployment, bootstrapping, and adding nodes to a cluster.
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

        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: "caaspctl/ci/infra/testrunner/testrunner stage=info ${PARAMS}", label: "Info")
            sh(script: "caaspctl/ci/infra/testrunner/testrunner stage=github_collaborator_check ${PARAMS}", label: "GitHub Collaborator Check")
            sh(script: "caaspctl/ci/infra/testrunner/testrunner stage=git_rebase branch-name=${env.BRANCH_NAME} ${PARAMS}", label: "Git Rebase Check")
            sh(script: "caaspctl/ci/infra/testrunner/testrunner stage=initial_cleanup ${PARAMS}", label: "Initial Cleanup")
        } }

        stage('Cluster Deployment') { steps {
            sh(script: "caaspctl/ci/infra/testrunner/testrunner stage=create_environment ${PARAMS}", label: "Create Environment")
            sh(script: "caaspctl/ci/infra/testrunner/testrunner stage=configure_environment ${PARAMS}", label: "Configure Environment")
        } }

        stage('Bootstrap Cluster') { steps {
            sh(script: "caaspctl/ci/infra/testrunner/testrunner stage=bootstrap_environment ${PARAMS}", label: "Bootstrap Environment")
        } }

        stage('Add Nodes in Cluster') { steps {
            sh(script: "caaspctl/ci/infra/testrunner/testrunner stage=grow_environment ${PARAMS}", label: "Grow Environment")
        } }
    }
    post {
        always {
            sh(script: "caaspctl/ci/infra/testrunner/testrunner stage=gather_logs ${PARAMS}", label: "Gather Logs")
            sh(script: "caaspctl/ci/infra/testrunner/testrunner stage=final_cleanup ${PARAMS}", label: "Final Cleanup")
            cleanWs()
        }
    }
}
