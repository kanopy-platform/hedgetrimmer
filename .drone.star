def main(ctx):
    platforms = [ "amd64","arm64"]

    volumes = [
        {
            "name": "cache",
            "path": "/go",
        }]

    workspace = {
        "path": "/go/src/github.com/${DRONE_REPO}"
    }

    resources = {
         "requests": {
              "cpu": 400,
              "memory": "2Gi"
         }
    }

    trigger = { "branch": ["main"] }

    pre_build_steps = {
        "test": append_volumes(test_step(), volumes),
        "license": append_volumes(license_step(), volumes),
    }
     
    pipelines = []
    for plat in platforms:
        pipe = { 
            "kind": "pipeline",
            "type": "kubernetes",
            "name": plat,
            "platform": { "arch": plat },
            "trigger": trigger,
            "steps": [ pre_build_steps.values() ]
        }

        bsnp = build(plat, False, False)
        bsnp = set_when(bsnp, {"event":["pull_request"]})
        bsnp = append_depends_on(bsnp, pre_build_steps.keys())
        bsnp = append_volumes(bsnp, volumes)

        bs = build(plat, False, True)
        bs = set_when(bs, {"event":["push"]})
        bs = append_depends_on(bs, pre_build_steps.keys())
        bs = append_volumes(bs, volumes)

        bstp = build(plat, True, True)
        bstp = set_when(bstp, {"event":["tag"]})
        bstp = append_depends_on(bstp, pre_build_steps.keys())
        bstp = append_volumes(bstp, volumes)
        pipe["steps"].append([bsnp, bs, bstp])

        pipelines.append(pipe)


    return pipelines


def build(arch, tag, publish):
    step = {
       "name": "publish",
       "image": "plugins/kaniko-ecr",
       "pull": "always",
       "environment": {
           "GIT_COMMIT": "${DRONE_COMMIT_SHA:0:7}",
       },
       "settings": {
            "repo": "${DRONE_REPO_NAME}",
            "build_args": [
                "GIT_COMMIT",
            ],
            "tags": [
                "git-${DRONE_COMMIT_SHA:0:7}-"+ arch,
            ],
       }
    }

    if tag:
        step["settings"]["tags"].append("${DRONE_TAG}-" + arch)
        step["environment"]["VERSION"] = "${DRONE_TAG}-" + arch
        step["settings"]["build_args"].append("VERSION")
    else:
        step["settings"]["tags"].append("latest-" + arch)

    if publish:
        step["settings"]["registry"] = { "from_secret": "ecr_registry" }
        step["settings"]["access_key"] = { "from_secret": "ecr_access_key" }
        step["settings"]["secret_key"] = { "from_secret": "ecr_secret_key" }
    else:
        step["settings"]["no_push"] = True
   
    return step

def test_step():
    return {
       "name": "test",
       "image": "golangci/golangci-lint:v1.49.0",
       "pull": "always",
       "commands": [ "make test" ]
    }

def license_step():
    return {
       "name": "license-check",
       "image": "public.ecr.aws/kanopy/licensed-go:3.7.3",
       "commands": ["licensed cache", "licensed status"]
    }
    

def set_when(step, when_condition):
    when_cons = getattr(step, "when", {})
    for k, v in when_condition.items():
        when_cons[k] = v

    step["when"] = when_cons
    return step

def append_volumes(step, vols):
    volumes = getattr(step, "volumes", [])
    for i in vols:
        volumes.append(i)

    step["volumes"] = volumes
    return step

def append_depends_on(step, refs):
    deps = getattr(step, "depends_on", [])

    for ref in refs:
        deps.append(ref)

    step["depends_on"] = deps
    return step

