/**
 * This pipeline verifies on a Github PR:
 *   - skuba unit and e2e tests
 *   - Basic skuba deployment, bootstrapping, and adding nodes to a cluster
 */

pipeline {
    agent { node { label 'caasp-team-private' } }

    environment {
        VMWARE_ENV_FILE = credentials('vmware-env')
        GITHUB_TOKEN = credentials('github-token')
        PLATFORM = 'vmware'
        TERRAFORM_STACK_NAME = "${JOB_NAME.replaceAll("/","-")}-${BUILD_NUMBER}"
        PR_CONTEXT = 'jenkins/skuba-test-vmware'
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
        REQUESTS_CA_BUNDLE = '/var/lib/ca-certificates/ca-bundle.pem'
        FILTER_SUBDIRECTORY = 'ci/infra/vmware'
        SHOULD_RUN = false
    }

    stages {
        stage('Setting GitHub in-progress status') { steps {
            sh(script: "${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'pending'", label: "Sending pending status")
        } }

        stage('Git Clone') { steps {
            deleteDir()
            checkout([$class: 'GitSCM',
                      branches: [[name: "*/${BRANCH_NAME}"], [name: '*/master']],
                      doGenerateSubmoduleConfigurations: false,
                      extensions: [[$class: 'LocalBranch'],
                                   [$class: 'WipeWorkspace'],
                                   [$class: 'RelativeTargetDirectory', relativeTargetDir: 'skuba']],
                      submoduleCfg: [],
                      userRemoteConfigs: [[refspec: '+refs/pull/*/head:refs/remotes/origin/PR-*',
                                           credentialsId: 'github-token',
                                           url: 'https://github.com/SUSE/skuba']]])

            dir("${WORKSPACE}/skuba") {
                sh(script: "git checkout ${BRANCH_NAME}", label: "Checkout PR Branch")
            }
        }}

        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f skuba/ci/Makefile pre_deployment', label: 'Pre Deployment')
            sh(script: 'make -f skuba/ci/Makefile pr_checks', label: 'PR Checks')
        } }

        stage('Check for changes') {
            steps {
                SHOULD_RUN = sh(script: "skuba/${PR_MANAGER} filter-pr --filename ${FILTER_SUBDIRECTORY}", returnStdout: true, label: "Filtering PR") =~ "contains changes"
            }
        }

        stage('Cluster Deployment') {
            when {
                expression { return SHOULD_RUN }
            }
            steps {
                sh(script: 'make -f skuba/ci/Makefile deploy', label: 'Deploy')
                archiveArtifacts("skuba/ci/infra/${PLATFORM}/terraform.tfstate")
                archiveArtifacts("skuba/ci/infra/${PLATFORM}/terraform.tfvars.json")
            }
        }

        stage('Run e2e tests') {
            when {
                expression { return SHOULD_RUN }
            }
            steps {
               sh(script: "make -f skuba/ci/Makefile test_e2e", label: "test_e2e")
            }
        }
    }
    post {
        always {
            script {
                if (SHOULD_RUN) {
                    sh(script: 'make --keep-going -f skuba/ci/Makefile post_run', label: 'Post Run')
                    zip(archive: true, dir: 'testrunner_logs', zipFile: 'testrunner_logs.zip')
                }
            }
        }
        cleanup {
            dir("${WORKSPACE}") {
                deleteDir()
            }
        }
        failure {
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'failure'", label: "Sending failure status")
        }
        success {
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'success'", label: "Sending success status")
        }
    }
}
