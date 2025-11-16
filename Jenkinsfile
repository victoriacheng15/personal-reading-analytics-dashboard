pipeline {
    agent any

    environment {
        GITHUB_USER = "victoriacheng15"
        GITHUB_TOKEN = credentials('GHCR_PAT')
        IMAGE_NAME = "ghcr.io/victoriacheng15/articles-extractor"
        IMAGE_TAG = "${env.BUILD_NUMBER}"
    }

    stages {
        stage('Clean Workspace') {
            steps {
                deleteDir()
            }
        }

        stage('Sanity Check') {
            steps {
                echo 'Starting Jenkins pipeline...'
            }
        }

        stage('Build Docker Image') {
            steps {
                script {
                    sh "docker build -t ${IMAGE_NAME}:${IMAGE_TAG} ."
                }
            }
        }

        stage('Login to GHCR') {
            steps {
                script {
                    sh "echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_USER --password-stdin"
                }
            }
        }

        stage('Push Image') {
            steps {
                script {
                    sh "docker push ${IMAGE_NAME}:${IMAGE_TAG}"
                    sh "docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${IMAGE_NAME}:latest"
                    sh "docker push ${IMAGE_NAME}:latest"
                }
            }
        }
    }

    post {
        always {
            sh "docker logout ghcr.io"
        }
    }
}
