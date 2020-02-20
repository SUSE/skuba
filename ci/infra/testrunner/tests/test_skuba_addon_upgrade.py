import pytest
import yaml
import re
import copy
from io import StringIO

def create_skuba_config(kubectl, configmap_data, dry_run=False):
    return kubectl.run_kubectl(
        'create configmap skuba-config --from-literal {0} -o yaml --namespace kube-system {1}'.format(
            configmap_data, '--dry-run' if dry_run else ''
        )
    )

def replace_skuba_config(kubectl, configmap_data):
    new_configmap = create_skuba_config(kubectl, configmap_data, dry_run=True)
    return kubectl.run_kubectl("replace -f -", stdin=new_configmap.encode())

def get_skubaConfiguration_dict(kubectl):
    skubaConf_yml = kubectl.run_kubectl(
        "get configmap skuba-config --namespace kube-system -o jsonpath='{.data.SkubaConfiguration}'"
    )
    return yaml.load(skubaConf_yml, Loader=yaml.FullLoader)

def decrease_one_addon_manifest(addons_dict, skip=None):
    for addon in addons_dict:
        if skip and addon in skip:
            continue
        ver = addons_dict[addon]['ManifestVersion']
        if ver > 0:
            addons_dict[addon]['ManifestVersion'] -= 1
            return (
                addon, addons_dict[addon]['Version'], (ver - 1, ver)
            )
    raise Exception("Could not decrease any addon manifest version!")

def change_one_addon_image(addons_dict, new_tag, skip=None):
    u_manif = decrease_one_addon_manifest(addons_dict, skip)
    addons_dict[u_manif[0]]['Version'] = new_tag
    return (u_manif[0], (new_tag, u_manif[1]), u_manif[2])

def remove_one_addon(addons_dict, skip=None):
    for addon in addons_dict.keys():
        if skip and addon in skip:
            continue
        version = addons_dict.pop(addon)
        return (
            addon, version['Version'], version['ManifestVersion']
        )
    raise Exception('Could not remove any addon!')

def addons_up_to_date(skuba):
    all_fine = re.compile(
        r'Congratulations! Addons for \d+\.\d+\.\d+ are already at the latest version available'
    )
    out = skuba.addon_upgrade('plan')
    return bool(all_fine.findall(out))


@pytest.mark.pr
def test_addon_upgrade_plan(deployment, kubectl, skuba):
    assert addons_up_to_date(skuba)

    skubaConf_dict = get_skubaConfiguration_dict(kubectl)
    skubaConf_orig = copy.deepcopy(skubaConf_dict)
    addons_dict = skubaConf_dict['AddonsVersion']

    rm_addon = remove_one_addon(addons_dict)
    u_manif = decrease_one_addon_manifest(addons_dict)
    u_img = change_one_addon_image(addons_dict, 'new_tag', skip=[u_manif[0]])

    u_img_msg = '{0}: {1} -> {2}'.format(u_img[0], u_img[1][0], u_img[1][1])
    u_manif_msg = '{0}: {1} -> {1} (manifest version from {2} to {3})'.format(
        u_manif[0], u_manif[1], u_manif[2][0], u_manif[2][1]
    )
    n_addon_msg = '{0}: {1} (new addon)'.format(rm_addon[0], rm_addon[1])

    replace_skuba_config(
        kubectl, "SkubaConfiguration='{0}'".format(yaml.dump(skubaConf_dict))
    )
    out = skuba.addon_upgrade('plan')
    replace_skuba_config(
        kubectl, "SkubaConfiguration='{0}'".format(yaml.dump(skubaConf_orig))
    )

    assert out.find(u_manif_msg) != -1
    assert out.find(u_img_msg) != -1
    assert out.find(n_addon_msg) != -1
    assert addons_up_to_date(skuba)


@pytest.mark.pr
def test_addon_upgrade_apply(deployment, kubectl, skuba):
    skubaConf_dict = get_skubaConfiguration_dict(kubectl)
    addons_dict = skubaConf_dict['AddonsVersion']

    decrease_one_addon_manifest(addons_dict)

    out = replace_skuba_config(
        kubectl, "SkubaConfiguration='{0}'".format(yaml.dump(skubaConf_dict))
    )
    assert not addons_up_to_date(skuba)

    out = skuba.addon_upgrade('apply')
    assert out.find("Successfully upgraded addons") != -1
    assert addons_up_to_date(skuba)
