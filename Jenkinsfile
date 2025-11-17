pipeline {

    agent none

    environment {
        GITHUB_USER = "victoriacheng15"
        GITHUB_TOKEN = credentials('GHCR_PAT')
        IMAGE_NAME = "ghcr.io/victoriacheng15/articles-extractor"
        IMAGE_TAG = "${env.BUILD_NUMBER}"
    }

    stages {

        stage('Clean & Checkout') {
            agent any
            steps {
                deleteDir()
                checkout scm
            }
        }

        stage('Code Formatting Check') {
            agent {
                docker {
                    image 'python:3.12-alpine'
                    args '-u root:root'
                }
            }
            steps {
                sh 'pip install ruff'

                sh '''
                    echo "Running ruff format check..."
                    if ruff format --check --diff main.py utils/ 2>/dev/null; then
                        echo "✓ Code is properly formatted"
                    else
                        echo "✗ Code formatting issues detected"
                        ruff format --check --diff main.py utils/
                        exit 1
                    fi
                '''
            }
        }

        stage('Build, Tag, Push (Docker)') {
            agent any
            steps {
                echo "Building Docker image..."
                sh "docker build -t ${IMAGE_NAME}:${IMAGE_TAG} ."

                echo "Logging into GHCR..."
                sh "echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_USER --password-stdin"

                echo "Pushing tagged image..."
                sh "docker push ${IMAGE_NAME}:${IMAGE_TAG}"

                echo "Updating latest tag..."
                sh "docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${IMAGE_NAME}:latest"
                sh "docker push ${IMAGE_NAME}:latest"

                echo "Logging out of GHCR..."
                sh "docker logout ghcr.io || true"
            }
        }
    }
}
