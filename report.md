
# Go Test Coverage Report

## Overall Coverage
- Total: ~90% (see package breakdown)

## Package Breakdown
- cmd/server: 0.0% (main function not unit testable; environment and compile tests included)
- internal/handlers: 97.1%
- internal/middleware: 93.5%
- internal/repository: 91.6%
- internal/services: 87.7%
- internal/utils: 100.0%

## Files Below Threshold (90%)
- internal/services/jwt_service.go (87.7%)
- internal/repository/token_store.go (91.6%)

## Recent Improvements
- Error-path tests for GenerateTokenPair in services
- All test files moved to correct package folders for coverage
- Import cycles and self-imports resolved

## Notes
- cmd/server/main.go excluded from coverage goals (server startup code, not unit testable)
- docs/swagger/docs.go excluded (generated code)
- internal/errors/errors.go excluded (constants only)
- internal/models/auth_model.go excluded (structs only)
- Target: Achieve ≥90% overall coverage through continued improvements