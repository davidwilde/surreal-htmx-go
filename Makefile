OUTPUT = app 
SOURCE_DIR = ./

# The default target to build and run the program
run: 
	go run $(SOURCE_DIR) 


# Watch for changes in the directory and rerun the program
watch:
	while true; do \
		go run $(SOURCE_DIR); \
		fswatch -r -1 $(SOURCE_DIR); \
	done

