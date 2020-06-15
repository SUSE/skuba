/**
 * This pipeline verifies on a Github PR:
 *   - skuba unit and e2e tests
 *   - Basic skuba deployment, bootstrapping, and adding nodes to a cluster
 */

// pr context to report
def pr_context = ''

// Platform for pr tests.
def platform = 'vmware'

// Repo branch
def branch_repo = ""

// CaaSP Version for repo branch
def repo_version = "4.0"

// type of worker required by the PR
def worker_type = 'integration'

// Set the agent platform label. Cannot be set using the environment variables
// because agent labels are evaluated before environment is set
def labels = ''
node('caasp-team-private-integration') {
    stage('select worker') {

        // If not a PR use BRANCH to select worker. Labels are not available. Skip rest of stage
        if (env.CHANGE_ID == null) {
            if (env.BRANCH_NAME.startsWith('experimental-') || env.BRANCH_NAME.startsWith('maintenance-')) {
                worker_type = env.BRANCH
            }
            currentBuild.result = 'SUCCESS'
            return
        }

        // check if PR needs experimental or maintenance workers
        // ci-worker label will override this selection
        if (env.CHANGE_TARGET.startsWith('experimental-') || env.CHANGE_TARGET.startsWith('maintenance-')) {
             worker_type = env.CHANGE_TARGET
        }

        try {
           def response = httpRequest(
               url: "https://api.github.com/repos/SUSE/skuba/pulls/${CHANGE_ID}",
               authentication: 'github-token',
               validResponseCodes: "200")

           def pr = readJSON text: response.content
           //check if the PR requires an specific worker type
           def pr_worker_label = pr.labels.find {
               it.name.startsWith("ci-worker:")
           }
           if (pr_worker_label != null) {
               worker_type = pr_worker_label.name.split(":")[1]
           }

          // check additional worker labels
           pr.labels.findAll {
             it.name.startsWith("ci-label:")
           }.each{
             def label = it.name.split(":")[1]
             labels = labels + " && " + label 
           }

           // check if the PR request an specific test platform
           def pr_platform_label = pr.labels.find {
               it.name.startsWith("ci-platform")
           }
           if (pr_platform_label != null) {
                platform = pr_platform_label.name.split(":")[1]
           }

           //check if the PR requires an specific worker type
           def pr_experimental_label = pr.labels.find {
               it.name.startsWith("ci-worker:")
           }
           if (pr_experimental_label != null) {
               worker_type = pr_experimental_label.name.split(":")[1]
           }

           //check if the PR requires an specific repository 
           def pr_repo_label = pr.labels.find {
               it.name.startsWith("ci-repo:")
           }
           if (pr_repo_label != null) {
               def branch_name = pr_repo_label.name.split(":")[1]
               branch_repo = "http://download.suse.de/ibs/Devel:/CaaSP:/${repo_version}:/Branches:/${branch_name}/SLE_15_SP1"

           }

        } catch (Exception e) {
            echo "Error retrieving labels for PR ${e.getMessage()}"
            currentBuild.result = 'ABORTED'
            error('Error retrieving labels for PR')
        }
    }
}

pipeline {
    agent { node { label "caasp-team-private-${worker_type} ${labels}" } }

    environment {
        SKUBA_BINPATH = '/home/jenkins/go/bin/skuba'
        VMWARE_ENV_FILE = credentials('vmware-env')
        GITHUB_TOKEN = credentials('github-token')
        PLATFORM = "${platform}" 
        TERRAFORM_STACK_NAME = "${BUILD_NUMBER}-${JOB_NAME.replaceAll("/","-")}".take(70)
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
        REQUESTS_CA_BUNDLE = '/var/lib/ca-certificates/ca-bundle.pem'
        LIBVIRT_URI = 'qemu+ssh://jenkins@kvm-ci.nue.caasp.suse.net/system'
        LIBVIRT_KEYFILE = credentials('libvirt-keyfile')
        LIBVIRT_IMAGE_URI = 'http://dist.suse.de/install/SLE-15-SP2-JeOS-PublicRC2/SLES15-SP2-JeOS.x86_64-15.2-OpenStack-Cloud-PublicRC2.qcow2'
    }

    stages {
        stage('Collaborator Check') { steps { script {
            pr_context = 'jenkins/skuba-validate-pr-author'
            sh(script: "${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${pr_context} 'pending'", label: "Sending pending status")

            if (env.BRANCH_NAME.startsWith('PR')) {
                def membersResponse = httpRequest(
                    url: "https://api.github.com/repos/SUSE/skuba/collaborators/${CHANGE_AUTHOR}",
                    authentication: 'github-token',
                    validResponseCodes: "204:404")

                if (membersResponse.status == 204) {
                    echo "Test execution for collaborator ${CHANGE_AUTHOR} allowed"
                    sh(script: "${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${pr_context} 'success'", label: "Sending success status")
                } else {
                    def allowExecution = false

                    try {
                        timeout(time: 5, unit: 'MINUTES') {
                            allowExecution = input(id: 'userInput', message: "Change author is not a SUSE member: ${CHANGE_AUTHOR}", parameters: [
                                booleanParam(name: 'allowExecution', defaultValue: false, description: 'Run tests anyway?')
                            ])
                        }
                    } catch(err) {
                        def user = err.getCauses()[0].getUser()
                        if('SYSTEM' == user.toString()) {
                            echo "Timeout while waiting for input"
                        } else {
                            allowExecution = false
                            echo "Unhandled error:\n${err}"
                        }
                    }

                    if (!allowExecution) {
                        echo "Test execution for unknown user (${CHANGE_AUTHOR}) disallowed"
                        error(message: "Test execution for unknown user (${CHANGE_AUTHOR}) disallowed")
                        return;
                    }
                }
            }
        } } }

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

        stage('code-lint') { steps { script {
            echo 'Starting code lint'
            pr_context = 'jenkins/skuba-code-lint'
            
            // set code lint status to pending
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${pr_context} 'pending'", label: "Sending pending status")

            dir("skuba") {
                sh(script: 'make lint', label: 'make lint')
            }

            echo 'Checking status of git tree'
            dir("skuba") {
                sh(script: 'test -z "$(git status --porcelain go.mod go.sum vendor/)" || { echo "there are uncommitted changes. This should never happen; diff:"; git diff; exit 1; }', label: 'git tree status')
            }

            echo 'Updating GitHub status for code-lint'
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${pr_context} 'success'", label: "Sending success status")

        } } }

        stage('Setting in-progress status for pr-test') { steps { script {
            pr_context = 'jenkins/skuba-test'
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${pr_context} 'pending'", label: "Sending pending status")
        } } }

        stage('Run skuba unit tests') { steps {
            dir("skuba") {
              sh(script: 'make test-unit', label: 'make test-unit')
            }
        } }

        stage('Getting Ready For Cluster Deployment') { 
            steps {
                sh(script: 'make -f skuba/ci/Makefile pre_deployment', label: 'Pre Deployment')
                sh(script: 'make -f skuba/ci/Makefile pr_checks', label: 'PR Checks')
                sh(script: "pushd skuba; make -f Makefile install; popd", label: 'Build Skuba')
            } 
        }

        stage('Provision cluster') {
            environment {
                BRANCH_REPO = "${branch_repo}"
            }
            steps {
                sh(script: 'make -f skuba/ci/Makefile provision', label: 'Provision')
            }
        }

        stage('Deploy cluster') {
            steps {
                sh(script: 'make -f skuba/ci/Makefile deploy', label: 'Deploy')
                sh(script: 'make -f skuba/ci/Makefile check_cluster', label: 'Check cluster')
            }
        }

        stage('Run e2e tests') { steps {
            sh(script: "make -f skuba/ci/Makefile test_pr", label: "test_pr")
        } }

        stage('Updating GitHub status for pr-test') { steps {
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${pr_context} 'success'", label: "Sending success status")
        } }

    }
    post {
        always { script {
            // collect artifacts only if pr-test stage was executed.
            if (pr_context == 'jenkins/pr-test'){
                archiveArtifacts(artifacts: "skuba/ci/infra/${PLATFORM}/terraform.tfstate", allowEmptyArchive: true)
                archiveArtifacts(artifacts: "skuba/ci/infra/${PLATFORM}/terraform.tfvars.json", allowEmptyArchive: true)
                archiveArtifacts(artifacts: 'testrunner.log', allowEmptyArchive: true)
                archiveArtifacts(artifacts: 'skuba/ci/infra/testrunner/*.xml', allowEmptyArchive: true)
                sh(script: "make --keep-going -f skuba/ci/Makefile gather_logs", label: 'Gather Logs')
                archiveArtifacts(artifacts: 'platform_logs/**/*', allowEmptyArchive: true)
                junit('skuba/ci/infra/testrunner/*.xml')
            }
        } }
        cleanup {
            sh(script: "make --keep-going -f skuba/ci/Makefile cleanup", label: 'Cleanup')
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
        unstable {
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${pr_context} 'failure'", label: "Sending failure status")
        }
        failure {
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${pr_context} 'failure'", label: "Sending failure status")
        }
        success {
            // status was alredy reported on each stage, no further action needed here
            echo "SUCCESS!"
        }
    }
}
