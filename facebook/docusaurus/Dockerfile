ARG TAG=18-alping
FROM node:$TAG

# Create Docusaurus directory and change working directory to that
RUN mkdir /docusaurus
WORKDIR /docusaurus

ARG SOURCE=./.posh/docs

ADD ${SOURCE}/package.json /docusaurus/
RUN npm install

ADD ${SOURCE}/ /docusaurus/

EXPOSE 3000
VOLUME [ "/docusaurus" ]
CMD [ "npm", "start" ]
