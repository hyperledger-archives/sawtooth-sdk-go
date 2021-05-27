#!groovy

// Copyright 2018 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ------------------------------------------------------------------------------

pipeline {
    agent {
        node {
            label 'master'
            customWorkspace "workspace/${env.BUILD_TAG}"
        }
    }

    triggers {
        cron(env.BRANCH_NAME == 'main' ? 'H 3 * * *' : '')
    }

    options {
        timestamps()
        buildDiscarder(logRotator(daysToKeepStr: '31'))
    }

    environment {
        ISOLATION_ID = sh(returnStdout: true, script: 'printf $BUILD_TAG | sha256sum | cut -c1-64').trim()
        COMPOSE_PROJECT_NAME = sh(returnStdout: true, script: 'printf $BUILD_TAG | sha256sum | cut -c1-64').trim()
    }

    stages {
        stage('Check User Authorization') {
            steps {
                readTrusted 'bin/authorize-cicd'
                sh './bin/authorize-cicd "$CHANGE_AUTHOR" /etc/jenkins-authorized-builders'
            }
            when {
                not {
                    branch 'main'
                }
            }
        }

        stage('Check for Signed-Off Commits') {
            steps {
                sh '''#!/bin/bash -l
                    if [ -v CHANGE_URL ] ;
                    then
                        temp_url="$(echo $CHANGE_URL |sed s#github.com/#api.github.com/repos/#)/commits"
                        pull_url="$(echo $temp_url |sed s#pull#pulls#)"
                        IFS=$'\n'
                        for m in $(curl -s "$pull_url" | grep "message") ; do
                            if echo "$m" | grep -qi signed-off-by:
                            then
                              continue
                            else
                              echo "FAIL: Missing Signed-Off Field"
                              echo "$m"
                              exit 1
                            fi
                        done
                        unset IFS;
                    fi
                '''
            }
        }

        stage('Fetch Tags') {
            steps {
                sh 'git fetch --tag'
            }
        }

        stage('Build Test Dependencies') {
            steps {
                sh 'docker build . -t sawtooth-sdk-go:$ISOLATION_ID'
                sh 'docker-compose -f docker-compose-installed.yaml build'
            }
        }

        stage('Run Lint') {
            steps {
                sh 'docker run --rm -v $(pwd):/go/src/github.com/hyperledger/sawtooth-sdk-go sawtooth-sdk-go:$ISOLATION_ID ./bin/run_go_fmt'
            }
        }

        stage('Run Tests') {
            steps {
                sh 'docker run --rm -v $(pwd):/go/src/github.com/hyperledger/sawtooth-sdk-go sawtooth-sdk-go:$ISOLATION_ID bash -c "cd tests && go test"'
                sh 'CORE=$(pwd) docker-compose -f tests/test_systemd_services.yaml up --abort-on-container-exit'
                sh 'docker-compose -f examples/intkey_go/tests/test_intkey_smoke_go.yaml up --abort-on-container-exit'
                sh 'docker-compose -f examples/intkey_go/tests/test_tp_intkey_go.yaml up --abort-on-container-exit'
                sh 'docker-compose -f examples/xo_go/tests/test_xo_smoke_go.yaml up --abort-on-container-exit'
                sh 'docker-compose -f examples/xo_go/tests/test_tp_xo_go.yaml up --abort-on-container-exit'
            }
        }

        stage("Build Archive artifacts") {
            steps {
                sh 'mkdir -p build/debs && docker-compose -f docker/compose/copy-debs.yaml up'
            }
        }
    }

    post {
        always {
            sh 'docker-compose -f examples/intkey_go/tests/test_intkey_smoke_go.yaml down'
            sh 'docker-compose -f examples/intkey_go/tests/test_tp_intkey_go.yaml down'
            sh 'docker-compose -f examples/xo_go/tests/test_xo_smoke_go.yaml down'
            sh 'docker-compose -f examples/xo_go/tests/test_tp_xo_go.yaml down'
            sh 'docker-compose -f docker/compose/copy-debs.yaml down'
        }
        success {
            archiveArtifacts 'build/debs/*.deb'
        }
        aborted {
            error "Aborted, exiting now"
        }
        failure {
            error "Failed, exiting now"
        }
    }
}
