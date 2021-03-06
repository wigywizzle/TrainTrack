# Use the official go docker image built on debian.
FROM golang:latest

# ENV VARS
ENV GOPATH $HOME/gocode
ENV PATH $HOME/gocode/bin:$PATH

# Grab the source code and add it to the workspace.
# ADD . /gocode/src/TrainTrack

# Install revel and the revel CLI.
RUN go get github.com/revel/revel
RUN go get github.com/revel/cmd/revel

# Install GoDep
RUN go get github.com/tools/godep


# # Grab the code from github and add it to the workspace #
RUN git clone https://github.com/williamczuba/TrainTrack.git $GOPATH/src/TrainTrack

# Change workdir for goDeps
WORKDIR $GOPATH/src/TrainTrack
RUN echo $PWD
RUN echo $(ls)

# Restore godep dependencies
RUN godep restore # $GOPATH/src/TrainTrack

# Use the revel CLI to start up our application.
# for google, should be 8080
WORKDIR $GOPATH
ENTRYPOINT revel run TrainTrack dev 8080 # Notice, we run in dev mode on port 8080

# Open up the port where the app is running.
# for google, should be 8080
EXPOSE 8080