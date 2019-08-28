/**
 * This pipeline verifies on a Github PR:
 *   - skuba unit and integration tests
 *   - Basic skuba deployment, bootstrapping, and adding nodes to a cluster
 */

pipeline {
    agent { node { label 'caasp-team-private' } }

    environment {
        SKUBA_BINPATH = '/home/jenkins/go/bin/skuba'
        OPENSTACK_OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        PLATFORM = 'openstack'
        TERRAFORM_STACK_NAME = "${JOB_NAME.replaceAll("/","-")}-${BUILD_NUMBER}"
        PR_CONTEXT = 'jenkins/skuba-test'
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
        REQUESTS_CA_BUNDLE = '/var/lib/ca-certificates/ca-bundle.pem'
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

        stage('Run skuba unit tests') { steps {
            dir("skuba") {
              sh(script: 'make test-unit', label: 'make test-unit')
            }
        } }

        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f skuba/ci/Makefile pre_deployment', label: 'Pre Deployment')
            sh(script: 'make -f skuba/ci/Makefile pr_checks', label: 'PR Checks')
            sh(script: "pushd skuba; make -f Makefile install; popd", label: 'Build Skuba')
        } }

        stage('Cluster Provisioning') {
            steps {
                sh(script: 'make -f skuba/ci/Makefile create_environment', label: 'Provision')
                archiveArtifacts(artifacts: "skuba/ci/infra/${PLATFORM}/terraform.tfstate", allowEmptyArchive: true)
                archiveArtifacts(artifacts: "skuba/ci/infra/${PLATFORM}/terraform.tfvars.json", allowEmptyArchive: true)
            }
        }

       stage('Run Pre Bootstrap Tests') {
           steps {
               sh(script: 'make -f skuba/ci/Makefile test_pre_bootstrap', label: 'Test Pre Bootstrap')
           }
       }

        stage('Cluster Bootstrap') {
            steps {
                sh(script: 'make -f skuba/ci/Makefile bootstrap', label: 'Bootstrap')
            }
        }

       stage('Run Post Bootstrap Tests') {
           steps {
               sh(script: 'make -f skuba/ci/Makefile test_post_bootstrap', label: 'Test Post Bootstrap')
           }
       }

        stage('Join Nodes') {
            steps {
                sh(script: 'make -f skuba/ci/Makefile join_nodes', label: 'Join Nodes')
            }
        }


        stage('Run Integration tests') { steps {
            sh(script: "make -f skuba/ci/Makefile test_integration", label: "test_integration")
        } }

    }
    post {
        always {
            junit('skuba/ci/infra/testrunner/*.xml')
            sh(script: "make --keep-going -f skuba/ci/Makefile post_run", label: 'Post Run')
            archiveArtifacts(artifacts: 'testrunner_logs/**/*', allowEmptyArchive: true)
            archiveArtifacts(artifacts: 'testrunner.log', allowEmptyArchive: true)
        }
        cleanup {
            dir("${WORKSPACE}@tmp") {
                deleteDir()
            }
            dir("${WORKSPACE}@script") {
                deleteDir()
            }
            dir("${WORKSPACE}@script@tmp") {
                deleteDir()
            }
            dir("${WORKSPACE}") {
                deleteDir()
            }
            sh(script: "rm -f ${SKUBA_BINPATH}; ", label: 'Remove built skuba')
        }
        failure {
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'failure'", label: "Sending failure status")
        }
        success {
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'success'", label: "Sending success status")
        }
    }
}
