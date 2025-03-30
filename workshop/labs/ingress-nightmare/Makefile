.PHONY: all clean

all: shared-libraries

shared-libraries:
	@echo "Compiling shared libraries..."
	gcc -fPIC -shared -o shared-library/hello_engine.so shared-library/hello_engine.c -lcrypto
	gcc -fPIC -shared -o shared-library/reverse_shell.so shared-library/reverse_shell.c -lcrypto
	@echo "Shared libraries compiled successfully"

clean:
	@echo "Cleaning up..."
	rm -f shared-library/*.so
	@echo "Clean complete"
