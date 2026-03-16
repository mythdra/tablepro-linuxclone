.PHONY: build clean run test

BUILD_DIR := build

build:
	cmake -B $(BUILD_DIR) -DCMAKE_TOOLCHAIN_FILE=/Users/can/vcpkg/scripts/buildsystems/vcpkg.cmake -DCMAKE_BUILD_TYPE=Debug
	cmake --build $(BUILD_DIR) --parallel

clean:
	rm -rf $(BUILD_DIR)

run:
	$(BUILD_DIR)/bin/TablePro.app/Contents/MacOS/TablePro

test:
	cmake --build $(BUILD_DIR) --target test_database_driver --parallel
	cmake --build $(BUILD_DIR) --target test_connection_manager --parallel
	$(BUILD_DIR)/bin/test_database_driver
	$(BUILD_DIR)/bin/test_connection_manager