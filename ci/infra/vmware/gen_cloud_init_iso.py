#!/usr/bin/env python3

# This script is called by Terraform to generate
# the cloud-init ISOs for the instances.

import os
import shutil
import subprocess
import sys
import tempfile


def get_command():
    try:
        if shutil.which("mkisofs"):
            return shutil.which("mkisofs")
        else:
            if shutil.which("genisoimage"):
                return shutil.which("genisoimage")
            else:
                raise
    except Exception:
        print("Please install mkisofs or genisoimage, aborting...")
        sys.exit(1)


def get_args():
    if len(sys.argv) <= 4:
        print("Please provide 4 arguments: $0 <worker|master|lb> "
              "<user_data> <meta_data> <net_config>")
        sys.exit(1)
    else:
        return sys.argv


def write_file(path, content):
    with open(path, "w", encoding="utf-8") as f:
        f.write(content)


def main():
    cmd = get_command()
    args = get_args()

    role = args[1]
    user_data = args[2]
    meta_data = args[3]
    net_config = args[4]
    iso_filename = "cc-{}.iso".format(role)
    script_path = os.path.dirname(os.path.abspath(__file__))

    if role not in ["worker", "master", "lb"]:
        print("The provided role <{}> is not correct, "
              "allowed values are <worker|master|lb>".format(role))
        sys.exit(1)

    # Create temporary directory
    tmp_dir = tempfile.mkdtemp()

    # Write the required files
    write_file("{0}/user-data".format(tmp_dir), user_data)
    write_file("{0}/meta-data".format(tmp_dir), meta_data)
    write_file("{0}/net-config".format(tmp_dir), net_config)

    # Generate ISO
    try:
        cmd_args = "{0} -output {1} -volid cidata -joliet -rock user-data meta-data net-config".format(cmd, iso_filename)
        subprocess.run(cmd_args, cwd=tmp_dir, shell=True, check=True)

        # Copy ISO to terraform directory
        shutil.copy(os.path.join(tmp_dir, iso_filename),
                    os.path.join(script_path, iso_filename))
    except subprocess.SubprocessError as e:
        print("Shelling out to {0} failed: {1}".format(cmd, e))

    # Clean temporary directory
    shutil.rmtree(tmp_dir, ignore_errors=True)


if __name__ == "__main__":
    main()
