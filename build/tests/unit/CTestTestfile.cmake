# CMake generated Testfile for 
# Source directory: /Users/can/code/tablepro-fork/tests/unit
# Build directory: /Users/can/code/tablepro-fork/build/tests/unit
# 
# This file includes the relevant testing commands required for 
# testing this directory and lists subdirectories to be tested as well.
add_test([=[test_database_driver]=] "/Users/can/code/tablepro-fork/build/bin/test_database_driver")
set_tests_properties([=[test_database_driver]=] PROPERTIES  _BACKTRACE_TRIPLES "/Users/can/code/tablepro-fork/tests/unit/CMakeLists.txt;12;add_test;/Users/can/code/tablepro-fork/tests/unit/CMakeLists.txt;0;")
add_test([=[test_connection_manager]=] "/Users/can/code/tablepro-fork/build/bin/test_connection_manager")
set_tests_properties([=[test_connection_manager]=] PROPERTIES  _BACKTRACE_TRIPLES "/Users/can/code/tablepro-fork/tests/unit/CMakeLists.txt;23;add_test;/Users/can/code/tablepro-fork/tests/unit/CMakeLists.txt;0;")
add_test([=[test_postgres_driver]=] "/Users/can/code/tablepro-fork/build/bin/test_postgres_driver")
set_tests_properties([=[test_postgres_driver]=] PROPERTIES  _BACKTRACE_TRIPLES "/Users/can/code/tablepro-fork/tests/unit/CMakeLists.txt;34;add_test;/Users/can/code/tablepro-fork/tests/unit/CMakeLists.txt;0;")
