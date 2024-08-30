import os
import socket
import struct

from nl80211 import Netlink

NETLINK_USERSPACE_PORT = 31  # Adjust as needed, matching the kernel module port

def create_netlink_socket():
  """Creates a netlink socket of type NETLINK_KOBJECT_UEVENT."""
  try:
    return Netlink(family=socket.AF_NETLINK, groups=socket.NETLINK_KOBJECT_UEVENT)
  except OSError as e:
    if e.errno == 95:  # Operation not supported (nl80211 might not be installed)
      print("Error: nl80211 library might not be installed. Communication might not work.")
      return None
    else:
      raise e

def send_message(sock, message):
  """Sends a message to the kernel module through the netlink socket."""
  msg = message.encode()
  nlmsg = struct.pack("HHII", len(msg) + 16, 0, 0, 0) + msg
  sock.send(nlmsg)

def receive_message(sock):
  """Receives a message from the kernel module through the netlink socket."""
  data = sock.recv(1024)
  (msglen,) = struct.unpack("I", data[:4])
  message = data[16:msglen]
  return message.decode()

def main():
  sock = create_netlink_socket()
  if not sock:
    return

  print("Netlink socket created and bound to port", NETLINK_USERSPACE_PORT)

  while True:
    print("\nAvailable options:")
    print("  1. Send message")
    print("  2. Receive message")
    print("  q. Quit")

    choice = input("Enter your choice: ")

    if choice.lower() == 'q':
      break
    elif choice == '1':
      message = input("Enter message to send: ")
      send_message(sock, message)
      print("Message sent to kernel module.")
    elif choice == '2':
      message = receive_message(sock)
      print("Received message from kernel module:", message)
    else:
      print("Invalid choice.")

  sock.close()

if __name__ == "__main__":
  main()

