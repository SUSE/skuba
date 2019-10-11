import pytest
import yaml
import re
from io import StringIO


def replace_skuba_config(kubectl, configmap_data):
    new_configmap = kubectl.run_kubectl(
        'create configmap skuba-config --from-literal {0} -o yaml --namespace kube-system --dry-run'.format(
            configmap_data
        )
    )
    return kubectl.run_kubectl("replace -f -", stdin=new_configmap.encode())

def get_skubaConfiguration_dict(kubectl):
    skubaConf_yml = kubectl.run_kubectl(
        "get configmap skuba-config --namespace kube-system -o jsonpath='{.data.SkubaConfiguration}'"
    )
    return yaml.load(skubaConf_yml, Loader=yaml.FullLoader)

def decrease_one_addon_manifest(addons_dict, skip=None):
    for addon in addons_dict:
        if addon == skip:
            continue
        ver = addons_dict[addon]['ManifestVersion']
        if ver > 0:
            addons_dict[addon]['ManifestVersion'] -= 1
            return (
                addon, addons_dict[addon]['Version'], (ver - 1, ver)
            )
    raise Exception("Could not decrease any addon manifest version!")

def decrease_one_addon_image(addons_dict, skip=None):
    for addon in addons_dict:
        if addon == skip:
            continue
        cur_ver = addons_dict[addon]['Version']
        addons_dict[addon]['Version'] = '0.0.1'
        return (
            addon, ('0.0.1', cur_ver), addons_dict[addon]['ManifestVersion']
        )
    raise Exception("Could not decrease any addon image version!")

def addons_up_to_date(skuba):
    all_fine = re.compile(
        r'Congratulations! Addons for \d+\.\d+\.\d+ are already at the latest version available'
    )
    out = skuba.addon_upgrade('plan')
    return bool(all_fine.findall(out))

def test_addon_upgrade_plan(deployment, kubectl, skuba):
    assert addons_up_to_date(skuba)

    skubaConf_dict = get_skubaConfiguration_dict(kubectl)
    addons_dict = skubaConf_dict['AddonsVersion']

    u_manif = decrease_one_addon_manifest(addons_dict)

    u_manif_msg = '{0}: {1} -> {1} (manifest version from {2} to {3})'.format(
        u_manif[0], u_manif[1], u_manif[2][0], u_manif[2][1]
    )

    out = replace_skuba_config(
        kubectl, "SkubaConfiguration='{0}'".format(yaml.dump(skubaConf_dict))
    )
    assert out.find("configmap/skuba-config replaced") != -1

    out = skuba.addon_upgrade('plan')
    assert out.find(u_manif_msg) != -1

def test_addon_upgrade_apply(deployment, kubectl, skuba):
    skubaConf_dict = get_skubaConfiguration_dict(kubectl)
    addons_dict = skubaConf_dict['AddonsVersion']

    decrease_one_addon_image(addons_dict)
    decrease_one_addon_manifest(addons_dict)

    out = replace_skuba_config(
        kubectl, "SkubaConfiguration='{0}'".format(yaml.dump(skubaConf_dict))
    )
    assert out.find("configmap/skuba-config replaced") != -1

    assert not addons_up_to_date(skuba)

    out = skuba.addon_upgrade('apply')
    assert out.find("Successfully upgraded addons") != -1
    assert addons_up_to_date(skuba)
