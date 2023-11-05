#!/usr/bin/env python3
import csv
import glob
import os
import shutil
import subprocess
import time

# List of tools and their parameters
tools_params = []

# ect will be used with --strict and -1 through -9
for i in range(9):
    tools_params.append(("ect", "--strict -" + str(i+1)))

# # oxipng will be used with -o 1 through 6
# for i in range(6):
#     tools_params.append(("oxipng", "-o " + str(i+1)))


tools = list({tool for tool, _ in tools_params})

not_installed = [tool for tool in tools if shutil.which(tool) is None]
if not_installed:
    print(f"The following tools are not installed or not in the PATH: {not_installed}")
    exit(1)


# The directory with the images that will be optimized
IMAGES_DIR = "test_files"

if not os.path.isdir(IMAGES_DIR):
    print(f"The directory {IMAGES_DIR} does not exist.")
    exit(1)

png_files = glob.glob(os.path.join(IMAGES_DIR, "*.png"))
if not png_files:
    print(f"The directory {IMAGES_DIR} does not contain any files with a .png extension.")
    exit(1)


RESULTS_FILE = "results.csv"



def run_benchmark(tool, params, image_path):
    # Record start time
    start = time.time()

    # Run the compression tool with parameters
    subprocess.run(f"{tool} {params} {image_path}", shell=True)

    # Record end time
    end = time.time()
    elapsed_time = end - start

    # Get output file size
    new_size = os.path.getsize(image_path)

    # Return results
    return new_size

def images_for_run(tool, params):
    directory = f"{tool}_{params}"

    shutil.copytree(IMAGES_DIR, directory)
    if not os.path.isdir(directory):
        print(f"The directory {directory} does not exist.")
        exit(1)

    files_to_optimize = glob.glob(os.path.join(IMAGES_DIR, "*.png"))
    if not files_to_optimize:
        print(f"The directory {directory} does not contain any files with a .png extension.")
        exit(1)

    return files_to_optimize

# Initialize CSV file with header
with open(RESULTS_FILE, mode='w', newline='') as file:
    writer = csv.writer(file)
    writer.writerow(["Tool", "Parameters", "InitialSize", "NewSize", "Ratio"])

for tool, params in tools_params:
    files_to_optimize = images_for_run(tool, params)
    for file_to_optimize in files_to_optimize:
        initial_size = os.path.getsize(file_to_optimize)
        new_size = run_benchmark(tool, params, file_to_optimize)
        ratio = new_size / initial_size
        with open(RESULTS_FILE, mode='a', newline='') as file:
            writer = csv.writer(file)
            writer.writerow((tool, params, initial_size, new_size, ratio))


# # Optionally, clean up the output images if not needed
# for tool, param in tools_params:
#     output_img_pattern = f"*_{tool}{param}.png"  # Or adjust the pattern as needed
#     for img in images:
#         try:
#             os.remove(os.path.join(IMAGES_DIR, output_img_pattern))
#         except OSError:
#             pass
