#include <sys/types.h>
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <string.h>
#include <errno.h>
#include <asm/types.h>
#include <linux/netlink.h>
#include <linux/socket.h>
#include <curl/curl.h>

#define NETLINK_USER 30 // same customized protocol as in my kernel module
#define MAX_PAYLOAD 1024 // maximum payload size
#define MAX_ID_LEN 100 // Adjust this based on your expected ID length
#define SERVER_URL "http://10.237.22.199:2020/get_file"

struct sockaddr_nl src_addr, dest_addr;
struct nlmsghdr *nlh = NULL;
struct nlmsghdr *nlh2 = NULL;
struct msghdr msg, resp;  // famous struct msghdr, it includes "struct iovec *   msg_iov;"
struct iovec iov, iov2;
int sock_fd;

// Function to read the unique ID from the file
int read_unique_id(const char *filename, char *id_buffer) {
  FILE *fp = fopen(filename, "r");
  if (fp == NULL) {
    return -1; // Error opening file
  }

  if (fgets(id_buffer, MAX_ID_LEN, fp) == NULL) {
    fclose(fp);
    return -2; // Error reading ID
  }

  // Remove trailing newline character
  size_t len = strlen(id_buffer);
  if (len > 0 && id_buffer[len - 1] == '\n') {
    id_buffer[len - 1] = '\0';
  }

  fclose(fp);
  return 0; // Success
}

// Callback function to write data received from the server
size_t writeCallback(void *contents, size_t size, size_t nmemb, void *userp) {
    size_t realsize = size * nmemb;
    char **response_ptr = (char **)userp;

    // Allocate memory for the response buffer
    *response_ptr = realloc(*response_ptr, realsize + 1);
    if (*response_ptr == NULL) {
        fprintf(stderr, "Failed to allocate memory for response buffer\n");
        return 0;
    }

    // Copy data from the server to the response buffer
    memcpy(*response_ptr, contents, realsize);
    (*response_ptr)[realsize] = '\0';

    return realsize;
}

// Function to download data from server using the ID
void download_data(const char *filename, const char *id) {
  CURL *curl;
  CURLcode res;
  char *buffer = NULL;
  long http_code = 0;
  size_t data_size = 0;

  curl = curl_easy_init();
  if (curl) {
    char url[strlen(SERVER_URL) + strlen(id) + 10]; // Allow space for ID and formatting
    sprintf(url, "%s/%s", SERVER_URL, id);

    curl_easy_setopt(curl, CURLOPT_URL, url);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, writeCallback); // No write function as we use a buffer
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &buffer); // Set buffer pointer for received data
    curl_easy_setopt(curl, CURLOPT_NOPROXY, "127.0.0.1");
    curl_easy_setopt(curl, CURLOPT_BUFFERSIZE, 67108864); // 64KB buffer

    // fprintf(stderr, "curl_easy_perform url: %s\n", url);
    // fprintf(stderr, "curl_easy_perform file: %s\n", filename);

    // Perform the request
    res = curl_easy_perform(curl);
    if (res != CURLE_OK) {
      fprintf(stderr, "curl_easy_perform failed: %s\n", curl_easy_strerror(res));
      goto cleanup;
    }

    // Write response data to file

    FILE *file = fopen(filename, "w");
    if (file == NULL) {
        fprintf(stderr, "Failed to open file for writing\n");
        goto cleanup;
    }

    fprintf(stderr, "The size of the buffer is %ld\n", strlen(buffer));

    if (fwrite(buffer, 1, strlen(buffer), file) != strlen(buffer)) {
        fprintf(stderr, "Failed to write data to file\n");
        fclose(file);
        goto cleanup;
    }
    fclose(file);

    printf("File content written successfully\n");

cleanup:
    curl_easy_cleanup(curl);
    free(buffer);

  }
}

int main(int args, char *argv[])
{
    //int socket(int domain, int type, int protocol);
    sock_fd = socket(PF_NETLINK, SOCK_RAW, NETLINK_USER); //NETLINK_KOBJECT_UEVENT  

    if(sock_fd < 0)
        return -1;

    memset(&src_addr, 0, sizeof(src_addr));
    src_addr.nl_family = AF_NETLINK;
    src_addr.nl_pid = getpid(); /* self pid */

    //int bind(int sockfd, const struct sockaddr *addr, socklen_t addrlen);
    if(bind(sock_fd, (struct sockaddr*)&src_addr, sizeof(src_addr))){
        perror("bind() error\n");
        close(sock_fd);
        return -1;
    }

    memset(&dest_addr, 0, sizeof(dest_addr));
    dest_addr.nl_family = AF_NETLINK;
    dest_addr.nl_pid = 0;       /* For Linux Kernel */

    //nlh: contains "Hello" msg
    nlh = (struct nlmsghdr *)malloc(NLMSG_SPACE(MAX_PAYLOAD));
    memset(nlh, 0, NLMSG_SPACE(MAX_PAYLOAD));
    nlh->nlmsg_len = NLMSG_SPACE(MAX_PAYLOAD);
    nlh->nlmsg_pid = getpid();  //self pid
    nlh->nlmsg_flags = 0; 

    //nlh2: contains received msg
    nlh2 = (struct nlmsghdr *)malloc(NLMSG_SPACE(MAX_PAYLOAD));
    memset(nlh2, 0, NLMSG_SPACE(MAX_PAYLOAD));
    nlh2->nlmsg_len = NLMSG_SPACE(MAX_PAYLOAD);
    nlh2->nlmsg_pid = getpid();  //self pid
    nlh2->nlmsg_flags = 0; 

    strcpy(NLMSG_DATA(nlh), "Done");   //put "Hello" msg into nlh

    iov.iov_base = (void *)nlh;         //iov -> nlh
    iov.iov_len = nlh->nlmsg_len;
    msg.msg_name = (void *)&dest_addr;  //msg_name is Socket name: dest
    msg.msg_namelen = sizeof(dest_addr);
    msg.msg_iov = &iov;                 //msg -> iov
    msg.msg_iovlen = 1;

    iov2.iov_base = (void *)nlh2;         //iov -> nlh2
    iov2.iov_len = nlh2->nlmsg_len;
    resp.msg_name = (void *)&dest_addr;  //msg_name is Socket name: dest
    resp.msg_namelen = sizeof(dest_addr);
    resp.msg_iov = &iov2;                 //resp -> iov
    resp.msg_iovlen = 1;

    int group = 17;
    if (setsockopt(sock_fd, SOL_NETLINK, NETLINK_ADD_MEMBERSHIP, &group, sizeof(group)) < 0) {
        printf(strerror(errno));
        close(sock_fd);
        return 1;
    }

    while (1)
    {
        // printf("Waiting for message from kernel\n");
        /* Read message from kernel */
        recvmsg(sock_fd, &resp, 0);  //msg is also receiver for read
        // printf("Received message payload: %s\n", (char *)NLMSG_DATA(nlh2));

        char full_path[1024];
        // strcpy(full_path, "/home"); // Copy base path FOR HDD
        strcpy(full_path, ""); // Copy base path FOR SSD
        strcat(full_path, (char *)NLMSG_DATA(nlh2));


        char id[MAX_ID_LEN];
        if (read_unique_id(full_path, id) != 0) {
            fprintf(stderr, "Error reading ID from file: %s\n", full_path);
            goto success;
        }

        // fprintf(stderr, "The ID read from the file: %s\n", id);
        char command[4096];
        sprintf(command, "curl --noproxy 10.237.22.199 -X GET %s/%s > %s ; tar -xzf %s ; cat %s > %s", SERVER_URL, id, id, id, id, full_path);

        int status = system(command);


        // download_data(full_path, id);

        // printf("Successfully replaced content of %s\n", full_path);

success:
        strcpy(NLMSG_DATA(nlh), "Done");   //put "Hello" msg into nlh
        // printf("Sending message \" %s \" to kernel\n");
        int ret = sendmsg(sock_fd, &msg, 0);   
        // printf("send ret: %d\n", ret);
        memset(nlh2, 0, NLMSG_SPACE(MAX_PAYLOAD));
    }

    close(sock_fd);

    return 0;
}

