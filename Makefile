.PHONY: build template clean link diff

# Build the templating program
build:
	go build -o vibeops main.go

# Run the templating process
template: build
	rm -rf build
	./vibeops template

# Run the templating process into prev-build
prev-build: build
	rm -rf prev-build
	./vibeops template -b prev-build

# Create symlinks from build directory to BaseDir
link: build
	./vibeops link

# Compare prev-build and build directories and restart changed services
diff: build
	./vibeops diff

# Clean up generated files
clean:
	rm -rf build vibeops
