CURRENT_DIR=$(shell pwd)
FUNCTION_PATH=$(shell basename ${CURRENT_DIR})
PREFIX=gitlab.udevs.io:5050/ucode_functions_group/${FUNCTION_PATH}
GATEWAY=https://ofs.u-code.io

gen-function:
	faas-cli new ${FUNCTION_PATH} --lang go --prefix ${PREFIX} --gateway ${GATEWAY}