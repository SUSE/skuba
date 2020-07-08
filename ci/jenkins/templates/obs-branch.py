#!/usr/bin/python3
import sh
from sh import osc
import lxml.builder
import lxml.etree
import os
import logging
import textwrap

def main():
  logging.basicConfig(level=logging.INFO)

  project=os.environ["PROJECT"]
  branch=os.environ["BRANCH"]
  requests=os.environ["REQUESTS"].split()
  rebuild=os.getenv("REBUILD", "containers")

  containers="Containers:CR"
  api = "https://api.suse.de"

  branchf = project+':'+branch
  cr_project = project+':'+containers
  cr_branch = branch+':'+containers
  cr_branchf = project+':'+cr_branch
  iosc = osc.bake("-A", api)
  iosc_post = iosc.api.bake('--method', 'POST')

  branch_exists = None
  try:
    iosc.api("/staging/{}/staging_projects/{}".format(project, branchf))
    branch_exists = True
  except sh.ErrorReturnCode_1:
    branch_exists = False

  cr_branch_exists = None
  try:
    iosc.api("/staging/{}/staging_projects/{}".format(project, cr_branchf))
    cr_branch_exists = True
  except sh.ErrorReturnCode_1:
    cr_branch_exists = False

  if not branch_exists:
    logging.info("Creating branch: %s", branchf)
    E = lxml.builder.ElementMaker()
    body = lxml.etree.tostring( E.workflow( E.staging_project( branchf ) ) )
    iosc.api('--method', 'POST', "/staging/{}/staging_projects".format(project), '-d', body)

    meta = iosc.meta.prj(branchf)
    xml_meta = lxml.etree.XML(str(meta))
    lxml.etree.strip_elements(xml_meta, "publish")
    repo = xml_meta.find(".//repository")
    if not repo:
      repo = E.repository(
        E.path(project=project, repository="SLE_15_SP2"),
        E.arch("x86_64"),
        E.arch("s390x"),
        E.arch("aarch64"),
        E.arch("ppc64le"),
        name="SLE_15_SP2")
      xml_meta.append(repo)
    meta = lxml.etree.tostring(xml_meta, pretty_print=True)
    logging.info("Setting meta of %s to %s", branchf, textwrap.indent(meta.decode("utf-8"), '    '))
    iosc.meta.prj('-F', '-', branchf, _in=meta)

    logging.info("Created branch: %s", branchf)

  if not cr_branch_exists:
    logging.info("Creating branch: %s", cr_branchf)
    E = lxml.builder.ElementMaker()
    body = lxml.etree.tostring( E.workflow( E.staging_project( cr_branchf ) ) )
    iosc.api('--method', 'POST', "/staging/{}/staging_projects".format(project), '-d', body)

    meta = iosc.meta.prj(cr_branchf)
    xml_meta = lxml.etree.XML(str(meta))
    lxml.etree.strip_elements(xml_meta, "publish")
    repo = xml_meta.find(".//repository")
    if not repo:
      repo = E.repository(
        E.path(project=cr_project, repository="containers"),
        E.arch("x86_64"),
        E.arch("s390x"),
        E.arch("aarch64"),
        E.arch("ppc64le"),
        name="containers")
      xml_meta.append(repo)
    repo.append( E.path(project="SUSE:SLE-15-SP2:Update:CR", repository="images") )
    repo.append( E.path(project=branchf, repository="SLE_15_SP2") )
    meta = lxml.etree.tostring(xml_meta, pretty_print=True)
    logging.info("Setting meta of %s to %s", cr_branchf, textwrap.indent(meta.decode("utf-8"), '    '))
    iosc.meta.prj('-F', '-', cr_branchf, _in=meta)

    kiwi = """
      %if "%_repository" == "containers"
      Type: kiwi
      Repotype: none
      Patterntype: none
      %endif
      """
    logging.info("Setting meta prjconf of %s to %s", cr_branchf, kiwi)
    kiwi = textwrap.dedent(kiwi)
    iosc.meta.prjconf('-F', '-', cr_branchf, _in=kiwi)

    logging.info("Created branch: %s", cr_branchf)

  if len(requests) > 0:
    package_requests = []
    container_requests = []
    for r in requests:
      out = iosc.api('/request/{}'.format(r))
      xml = lxml.etree.XML(str(out))
      targets = xml.xpath("./action/target/@project")
      if len(targets) > 1:
        logging.warning("Please split request %s up, it has multiple actions.", r)
        continue
      elif len(targets) < 1:
        logging.warning("Skipping request %s, as it has not actions with a target.", r)
        continue
      target = targets[0]
      if target == cr_project:
        container_requests.append(r)
      elif target == project:
        package_requests.append(r)
      else:
        logging.warning("Skipping request %s, as its target %s matches neither the project nor the container project.", r, target)

    if len(package_requests) > 0:
      E = lxml.builder.ElementMaker()
      xml_requests = E.requests()
      for r in package_requests:
        xml_requests.append(E.request(id=r))
      body = lxml.etree.tostring( xml_requests )
      logging.info("Adding to branch %s: %s", branchf, body)
      iosc_post("/staging/{}/staging_projects/{}/staged_requests".format(project, branchf), '-d', body)

    if len(container_requests) > 0:
      xml_requests = E.requests()
      for r in package_requests:
        xml_requests.append(E.request(id=r))
      body = lxml.etree.tostring( xml_requests )
      logging.info("Adding to branch %s: %s", cr_branchf, body)
      iosc_post("/staging/{}/staging_projects/{}/staged_requests".format(project, cr_branchf), '-d', body)

  if rebuild in ["containers", "both"]:
    done = list(iosc.ls(cr_branchf, _iter=True))
    for line in iosc.ls(cr_project, _iter=True):
      if line in done:
        logging.info("Package %s alread exists in %s", line.strip(), cr_branchf)
        continue
      line = line.strip()
      logging.info("Linking %s into %s", line, cr_branchf)
      iosc.linkpac(cr_project, line, cr_branchf)

  if rebuild in ["packages", "both"]:
    done = list(iosc.ls(branchf, _iter=True))
    for line in iosc.ls(project, _iter=True):
      if line in done:
        logging.info("Package %s alread exists in %s", line.strip(), branchf)
        continue
      line = line.strip()
      logging.info("Linking %s into %s", line, branchf)
      iosc.linkpac(cr_project, line, branchf)

if __name__ == "__main__":
  main()
