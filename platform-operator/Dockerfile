# Copyright (c) 2020, 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

ARG BASE_IMAGE=ghcr.io/oracle/oraclelinux:8-slim
FROM $BASE_IMAGE AS build_base

# Need to use specific WORKDIR to match verrazzano's source packages
WORKDIR /root/go/src/github.com/verrazzano/verrazzano/platform-operator
COPY . .

COPY out/linux_amd64/verrazzano-platform-operator /usr/local/bin/verrazzano-platform-operator

RUN chmod 500 /usr/local/bin/verrazzano-platform-operator \
    && chmod +x scripts/install/*.sh \
    && chmod +x scripts/*.sh
# Create the verrazzano-platform-operator image
FROM $BASE_IMAGE AS final

ARG VERRAZZANO_PLATFORM_OPERATOR_IMAGE

# copy olcne repos needed to install kubectl, istioctl
COPY --from=build_base /root/go/src/github.com/verrazzano/verrazzano/platform-operator/repos/*.repo /etc/yum.repos.d/

RUN microdnf install -y --setopt=install_weak_deps=0 --setopt=tsflags=nodocs openssl jq kubectl-1.23.11-1.el8 istio-istioctl-1.15.3-2.el8 \
    && microdnf clean all \
    && rm -rf /var/cache/yum /var/lib/rpm/* \
    && groupadd -r verrazzano \
    && useradd --no-log-init -r -m -d /verrazzano -g verrazzano -u 1000 verrazzano \
    && mkdir /home/verrazzano \
    && chown -R 1000:verrazzano /home/verrazzano

# Copy the operator binary
COPY --from=build_base --chown=verrazzano:verrazzano /usr/local/bin/verrazzano-platform-operator /usr/local/bin/verrazzano-platform-operator

# Copy the Verrazzano install and uninstall scripts
WORKDIR /verrazzano
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano/platform-operator/thirdparty ./platform-operator/thirdparty
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano/platform-operator/manifests ./platform-operator/manifests
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano/platform-operator/scripts/wait4webhook.sh ./platform-operator/scripts/wait4webhook.sh
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano/platform-operator/scripts/hooks ./platform-operator/scripts/hooks
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano/platform-operator/scripts/install ./platform-operator/scripts/install
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano/platform-operator/config/scripts/run.sh .
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano/platform-operator/config/scripts/kubeconfig-template ./config/kubeconfig-template
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano/platform-operator/helm_config ./platform-operator/helm_config
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano/platform-operator/out/generated-verrazzano-bom.json ./platform-operator/verrazzano-bom.json
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano/platform-operator/capi/bootstrap-ocne ./capi/bootstrap-ocne
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano/platform-operator/capi/control-plane-ocne ./capi/control-plane-ocne
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano/platform-operator/capi/infrastructure-oci ./capi/infrastructure-oci
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano/platform-operator/capi/cluster-api ./capi/cluster-api

COPY --from=build_base /root/go/src/github.com/verrazzano/verrazzano/platform-operator/THIRD_PARTY_LICENSES.txt /licenses/

# set the default VPO image in values.yaml for the VPO helm chart
RUN  sed -i -e "s|image:|image: $VERRAZZANO_PLATFORM_OPERATOR_IMAGE|g" ./platform-operator/helm_config/charts/verrazzano-platform-operator/values.yaml

USER 1000

ENTRYPOINT ["/verrazzano/run.sh"]
