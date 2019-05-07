#!/usr/bin/env python3

import shutil
import subprocess
import sys
import tempfile


def write_file(path, content):
    with open(path, "w", encoding="utf-8") as f:
        f.write(content)


def main():
    # Check commands exists
    if shutil.which("mkisofs"):
        cmd = shutil.which("mkisofs")
    else:
        if shutil.which("genisoimage"):
            cmd = shutil.which("genisoimage")
        else:
            print("Please install mkisofs or genisoimage, aborting...")
            sys.exit()

    # Validate arguments
    if len(sys.argv) <= 4:
        print("Please provide 4 arguments: $0 <worker|master|lb> "
              " <user_data> <meta_data> <net_config>")
        sys.exit()
    else:
        role = sys.argv[1]
        user_data = sys.argv[2]
        meta_data = sys.argv[3]
        net_config = sys.argv[4]
        iso_filename = "cc-{}.iso".format(role)

    if role not in ["worker", "master", "lb"]:
        print("The provided role <{}> is not correct, "
              "allowed values are <worker|master|lb>".format(role))
        sys.exit()

    # Create temporary directory
    tmp_dir = tempfile.mkdtemp()

    # Write the required files
    write_file("{0}/user-data".format(tmp_dir), user_data)
    write_file("{0}/meta-data".format(tmp_dir), meta_data)
    write_file("{0}/net-config".format(tmp_dir), net_config)

    # Create ISO
    try:
        args = "{0} -output {1} -volid cidata -joliet -rock user-data meta-data net-config".format(cmd, iso_filename)
        subprocess.run(args, cwd=tmp_dir, shell=True, check=True)
    except subprocess.SubprocessError as e:
        print("Shelling out to {0} failed: {1}".format(cmd, e))

    # Clean temporary directory
    if tmp_dir:
        shutil.rmtree(tmp_dir)


if __name__ == "__main__":
    main()
