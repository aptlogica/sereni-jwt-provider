pipeline {
  agent any

  stages {
    stage('Checkout Code') {
      steps {
        checkout scm
      }
    }

    stage('Test & Coverage') {
      steps {
        sh '''
          go test -v -race -coverprofile=coverage.out -covermode=atomic \
            -coverpkg=github.com/aptlogica/sereni-jwt-provider,github.com/aptlogica/sereni-jwt-provider/internal/config,github.com/aptlogica/sereni-jwt-provider/internal/handlers,github.com/aptlogica/sereni-jwt-provider/internal/services,github.com/aptlogica/sereni-jwt-provider/internal/utils \
            ./tests/...
          go tool cover -func=coverage.out
        '''
      }
    }

    stage('SonarQube Analysis') {
        when {
        anyOf {
          branch 'develop'
          branch 'main'
          branch 'release/*'
          branch 'master'
        }
      }
      steps {
        script {
          // Get path to the installed Sonar Scanner tool
          def scannerHome = tool 'SonarScanner'

          withSonarQubeEnv('aptl-sonar') {
            // Run the scanner binary
            sh "${scannerHome}/bin/sonar-scanner"
          }
        }
      }
    }

    stage('Quality Gate') {
      when {
        anyOf {
          branch 'develop'
          branch 'main'
          branch 'release/*'
          branch 'master'
        }
      }
      steps {
        timeout(time: 10, unit: 'MINUTES') {
          waitForQualityGate abortPipeline: true
        }
      }
    }
  }
}
