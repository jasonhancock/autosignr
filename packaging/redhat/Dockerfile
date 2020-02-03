FROM centos:7
ARG goversion

RUN yum -y install \
  gcc \
  git \
  make \
  rpm-build \
  rsync \
  wget

ENV gofile go${goversion}.linux-amd64.tar.gz
ENV gourl https://storage.googleapis.com/golang/${gofile}

RUN wget -q -O /usr/local/${gofile} ${gourl}
RUN mkdir /usr/local/go
RUN tar -xzf /usr/local/${gofile} -C /usr/local/go --strip 1

CMD cd /mnt/build && PATH=/usr/local/go/bin:$PATH make package

