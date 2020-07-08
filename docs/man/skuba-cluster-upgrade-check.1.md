% skuba-cluster-upgrade-check(1) # skuba cluster upgrade check - Returns the list of deprecated APIs used by skuba

# NAME
check - Lists deprecated APIs used

# SYNOPSIS
**check**
[**--api-walk**] [**--description**] [**--kubernetes-version**]
[**--swagger-dir**] [**--force-download**] [**--format**]
[**--filename**] [**--input-file**] [**--help**|**-h**]
*check* [kubernetes-version=*<version>*] [swaggerDir=*<directory>*] [--api-walk=*<true|fasle>*] [--description] [--force-download] [--input-file=*<filename>*]

# DESCRIPTION
**check** list deprecated APIs used for the next upgradable version of skuba.

# OPTIONS

**--help, -h**
  Print usage statement.

**api-walk**
  Whether to walk in the whole API, checking if all objects type still exists in the current swagger.json. May be IO intensive to APIServer.

**description**
  Whether to show the description of the deprecated object. The description may contain the solution for the deprecation. (default true)

**kubernetes-version**
  Which kubernetes release version (https://github.com/kubernetes/kubernetes/releases) should be used to validate objects. (Default=next upgradable kubernetes version)

**swagger-dir**
  Where to keep swagger.json downloaded file. If not provided will use the system temporary directory.

**force-download**
  Whether to force the download of a new swagger.json file even if one exists.

**format**
  Format in which the list will be displayed [stdout, plain, json, yaml]. (default plain).

**filename**
  Name of the file the results will be saved to, if empty it will display to stdout.

**input-file**
  Location of a file or directory containing kubernetes manifests to be analized.