pipeline {
    agent any

    environment {
        GITHUB_USER = "victoriacheng15"
        GITHUB_TOKEN = credentials('GHCR_PAT')
    }

    stages {
        stage('Sanity Check') {
            steps {
                echo 'The beginning of Jenkinsfile'
            }
        }
    }
}