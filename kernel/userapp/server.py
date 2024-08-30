from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route("/get_file/<id>", methods=["GET"])
def get_file(id):
    """Retrieves data from the server based on the provided ID."""
    try:
        with open(f"./compressed_data/{id}", 'rb') as file:
            file_content = file.read()
        return file_content
    except Exception as e:
        return "YOU GOT REPLACED"

if __name__ == "__main__":
  app.run(host="0.0.0.0", debug=True)  # Adjust host and debug as needed
