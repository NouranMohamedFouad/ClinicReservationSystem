FROM node:14
# Set the working directory
WORKDIR /app

RUN chmod 777 -R /app
# Add the node_modules bin to the PATH
ENV PATH /app/node_modules/.bin:$PATH
# Copy package.json and package-lock.json
COPY package.json /app/
COPY package-lock.json /app/

# Install npm dependencies
RUN npm install
# Install react-scripts globally
RUN npm install react-scripts@5.0.1 -g
# Copy the application code
COPY . /app/

EXPOSE 3000
# Set the command to start the application
ENTRYPOINT [ "npm", "start" ]
