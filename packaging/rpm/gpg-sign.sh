#!/usr/bin/env bash

set -euxo pipefail
shopt -s extglob

# We are not using fpm's signing functionality because it does not work anymore
# https://github.com/jordansissel/fpm/issues/1626

which gpg

# Write GPG key to GPG keyring
printf "%s" "${GPG_PUBLIC_KEY}" > /tmp/gpg-public-key
gpg --import /tmp/gpg-public-key
printf "%s" "${GPG_PRIVATE_KEY}" | gpg --import --no-tty --batch --yes --passphrase "${GPG_PASSPHRASE}"

rpm --import /tmp/gpg-public-key

echo "%_gpg_name Grafana <info@grafana.com>
%_signature gpg
%_gpg_path /root/.gnupg
%_gpgbin /usr/bin/gpg
%__gpg_check_password_cmd /bin/true
%__gpg_sign_cmd     %{__gpg} \
         gpg --no-tty --batch --yes --verbose --no-armor \
         --passphrase "${GPG_PASSPHRASE}" \
         --pinentry-mode loopback \
         %{?_gpg_digest_algo:--digest-algo %{_gpg_digest_algo}} \
         --no-secmem-warning \
         -u \"%{_gpg_name}\" -sbo %{__signature_filename} %{__plaintext_filename}
" > ~/.rpmmacros

cat /dev/null | setsid rpmsign --resign dist/*.rpm
rpm --checksig dist/*.rpm
