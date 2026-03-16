# CMakeLists.txt for integration tests

# Integration test for PostgresDriver with actual PostgreSQL database
set(TEST_POSTGRES_INTEGRATION_SOURCES
    test_postgres_driver_integration.cpp
)

add_executable(test_postgres_integration ${TEST_POSTGRES_INTEGRATION_SOURCES})
target_link_libraries(test_postgres_integration Qt6::Core Qt6::Test tablepro_core tablepro_drivers)
target_include_directories(test_postgres_integration PRIVATE ${CMAKE_SOURCE_DIR}/include)

add_test(NAME test_postgres_integration COMMAND test_postgres_integration)