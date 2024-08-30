import os
import tarfile

def create_individual_tar(directory):
  for root, _, files in os.walk(directory):
    for filename in files:
      filepath = os.path.join(root, filename)
      # Create archive filename with .tar.gz extension
      archive_name = f"{filename}"
      archive_path = os.path.join("/home/user/FaaSSnapper/userapp/compressed_data", archive_name)

      # Create a new tar archive with gzip compression
      with tarfile.open(archive_path, mode="w:gz") as tar:
        # Add the single file to the archive
        tar.add(filepath, arcname=filename)
      print(f"Created archive: {archive_path}")

if __name__ == "__main__":
  create_individual_tar("/home/user/FaaSSnapper/userapp/data")
