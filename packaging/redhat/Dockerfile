FROM centos:7

RUN yum -y install \
  git \
  make \
  rpm-build \
  rsync \
  wget

ENV goversion 1.11.5
ENV gofile go${goversion}.linux-amd64.tar.gz
ENV gourl https://storage.googleapis.com/golang/${gofile}

RUN wget -q -O /usr/local/${gofile} ${gourl}
RUN mkdir /usr/local/go
RUN tar -xzf /usr/local/${gofile} -C /usr/local/go --strip 1

CMD cd /mnt/build && PATH=/usr/local/go/bin:$PATH make package

