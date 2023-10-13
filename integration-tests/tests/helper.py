import os
import re

def replace_string_in_file(filename, pattern, replacement):
    with open(filename, 'r') as file:
        filedata = file.read()

    new_data = re.sub(pattern, replacement, filedata)

    with open(filename, 'w') as file:
        file.write(new_data)

def main():
    base_dir = os.path.dirname(os.path.abspath(__file__))

    # Renaming the folders first
    for folder in os.listdir(base_dir):
        if folder.startswith("scrap-prom-metrics copy"):
            new_folder_name = folder.replace(" ", "-")
            os.rename(os.path.join(base_dir, folder), os.path.join(base_dir, new_folder_name))

    # Processing the renamed folders
    for folder in os.listdir(base_dir):
        if folder.startswith("scrap-prom-metrics-copy"):
            suffix_number = re.search(r'(\d+)$', folder)
            if suffix_number:
                suffix_number = suffix_number.group(1)
                
                # Replace in scrap_prom_metrics_test.go
                go_file_path = os.path.join(base_dir, folder, "scrap_prom_metrics_test.go")
                if os.path.exists(go_file_path):
                    replace_string_in_file(
                        go_file_path,
                        r"avalanche_metric_mmmmm_0_0",
                        f"avalanche_metric_mmmmm_0_{suffix_number}"
                    )
                    replace_string_in_file(
                        go_file_path,
                        r"scrap_prom_metrics",
                        f"scrap_prom_metrics{suffix_number}"
                    )
                
                # Replace in config.river
                river_file_path = os.path.join(base_dir, folder, "config.river")
                if os.path.exists(river_file_path):
                    replace_string_in_file(
                        river_file_path,
                        r"scrap_prom_metrics",
                        f"scrap_prom_metrics{suffix_number}"
                    )

if __name__ == "__main__":
    main()
