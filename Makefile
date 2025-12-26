.PHONY: build template clean link

# Build the templating program
build:
	go build -o vibeops main.go

# Run the templating process
template: build
	./vibeops template

# Create symlinks from build directory to BaseDir
link: build
	./vibeops link

# Clean up generated files
clean:
	rm -rf build vibeops
