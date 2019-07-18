pipeline {
    agent { node { label 'caasp-team-private' } }

    environment {
        OPENSTACK_OPENRC = credentials('ecp-openrc')
        TERRAFORM_STACK_NAME = "${JOB_NAME.replaceAll("/","-")}-${BUILD_NUMBER}"
        GITHUB_TOKEN = credentials('github-token')
        VMWARE_ENV_FILE = credentials('vmware-env')
    }

    stages {
        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f skuba/ci/Makefile pre_deployment', label: 'Pre Deployment')
        } }

        stage('Cluster Deployment') { steps {
            sh(script: 'make -f skuba/ci/Makefile deploy', label: 'Deploy')
            archiveArtifacts("skuba/ci/infra/${PLATFORM}/terraform.tfstate")
            archiveArtifacts("skuba/ci/infra/${PLATFORM}/terraform.tfvars.json")
        } }

        stage('Bootstrap Cluster') { steps {
            sh(script: 'make -f skuba/ci/Makefile bootstrap', label: 'Bootstrap')
            sh(script: "skuba/ci/infra/testrunner/testrunner --platform ${PLATFORM} join-node --role worker --node 0", label: 'Join Worker 0')
            sh(script: "skuba/ci/infra/testrunner/testrunner --platform ${PLATFORM} join-node --role worker --node 1", label: 'Join Worker 1')
        } }

        stage('Conformance Tests') { steps {
            sh(script: "skuba/ci/tasks/sonobuoy_e2e.py run --kubeconfig ${WORKSPACE}/test-cluster/admin.conf", label: 'Run Conformance')
            sh(script: "skuba/ci/tasks/sonobuoy_e2e.py collect --kubeconfig ${WORKSPACE}/test-cluster/admin.conf", label: 'Collect Results')
            archiveArtifacts('results')
            junit('results/plugins/e2e/results/*.xml')
            sh(script: "skuba/ci/tasks/sonobuoy_e2e.py cleanup --kubeconfig ${WORKSPACE}/test-cluster/admin.conf", label: 'Cleanup Cluster')
        } }

    }
    post {
        always {
            sh(script: 'make --keep-going -f skuba/ci/Makefile post_run', label: 'Post Run')
            zip(archive: true, dir: 'testrunner_logs', zipFile: 'testrunner_logs.zip')
        }
        cleanup {
            dir("${WORKSPACE}") {
                deleteDir()
            }
        }
    }
}
