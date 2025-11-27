pipeline {
    agent none

    environment {
        GITHUB_USER   = "victoriacheng15"
        GITHUB_TOKEN  = credentials('GHCR_PAT')
        IMAGE_NAME    = "ghcr.io/victoriacheng15/articles-extractor"
        IMAGE_TAG     = ""  // set dynamically
    }

    stages {
        stage('Prepare Workspace') {
            agent any
            steps {
                deleteDir()
                script {
                    env.IMAGE_TAG = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                }
                echo "✅ Using Git SHA as image tag: ${env.IMAGE_TAG}"
            }
        }

        stage('Code Formatting Check') {
            agent {
                docker {
                    image 'python:3.12-alpine'
                    args '-u 1000:1000'
                }
            }
            environment {
                HOME = '/tmp'
            }
            steps {
                sh 'pip install --no-warn-script-location ruff'
                sh '''
                    RUFF="/tmp/.local/bin/ruff"
                    echo "Running ruff format check..."
                    if "$RUFF" format --check --diff main.py utils/ 2>/dev/null; then
                        echo "✓ Code is properly formatted"
                    else
                        echo "✗ Code formatting issues detected"
                        "$RUFF" format --check --diff main.py utils/
                        exit 1
                    fi
                '''
            }
        }

        stage('Build, Tag, Push (Docker)') {
            agent any
            steps {
                sh "docker build -t ${IMAGE_NAME}:${IMAGE_TAG} ."

                sh "echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_USER --password-stdin"

                sh "docker push ${IMAGE_NAME}:${IMAGE_TAG}"

                // Optional: push :latest (consider only on main branch in future)
                sh "docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${IMAGE_NAME}:latest"
                sh "docker push ${IMAGE_NAME}:latest"

                sh "docker logout ghcr.io || true"
            }
        }
    }
}