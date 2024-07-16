#!/bin/sh
set -e

# compile for version
make
if [ $? -ne 0 ]; then
   echo "make error"
   exit 1
fi

tun_version=`./bin/tuns --version`
echo "build version: $tun_version"

# cross_compiles
rm -rf ./release

make -f ./Makefile.cross-compiles

rm -rf ./release/packages
mkdir -p ./release/packages

os_all='linux windows darwin freebsd android'
arch_all='386 amd64 arm arm64 mips64 mips64le mips mipsle riscv64'
extra_all='_ hf'

cd ./release
for os in $os_all; do
    for arch in $arch_all; do
        for extra in $extra_all; do
            suffix="${os}_${arch}"
            if [ "x${extra}" != x"_" ]; then
                suffix="${os}_${arch}_${extra}"
            fi
            tun_dir_name="tun_${tun_version}_${suffix}"
            tun_path="./packages/tun_${tun_version}_${suffix}"

            if [ "x${os}" = x"windows" ]; then
                if [ ! -f "./tunc_${os}_${arch}.exe" ]; then
                    continue
                fi
                if [ ! -f "./tuns_${os}_${arch}.exe" ]; then
                    continue
                fi
                mkdir ${tun_path}
                mv ./tunc_${os}_${arch}.exe ${tun_path}/tunc.exe
                mv ./tuns_${os}_${arch}.exe ${tun_path}/tuns.exe
            else
                if [ ! -f "./tunc_${suffix}" ]; then
                    continue
                fi
                if [ ! -f "./tuns_${suffix}" ]; then
                    continue
                fi
                mkdir ${tun_path}
                mv ./tunc_${suffix} ${tun_path}/tunc
                mv ./tuns_${suffix} ${tun_path}/tuns
            fi

            # packages
            cd ./packages
            if [ "x${os}" = x"windows" ]; then
                zip -rq ${tun_dir_name}.zip ${tun_dir_name}
            else
                tar -zcf ${tun_dir_name}.tar.gz ${tun_dir_name}
            fi
            cd ..
            rm -rf ${tun_path}
        done
    done
done

cd -

