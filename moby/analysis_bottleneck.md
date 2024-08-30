Analysis

Hunch: The image is the culprit in the snapshot process. Let us dig on the image handler of Docker.

Detailed explanation:
"github.com/containerd/containerd/log"
log.L.Errorf("FAAS: FAAS: FAAS \n")

Bottleneck-I:
FAAS: The container checkpoint task took 1.296964045s seconds 

Starting point: vendor/github.com/containerd/containerd/task.go in function Checkpoint()
'''
	index := v1.Index{
		Versioned: is.Versioned{
			SchemaVersion: 2,
		},
		Annotations: make(map[string]string),
	}

	if err := t.checkpointTask(ctx, &index, request); err != nil {
		return nil, err
	}
'''
The checkpoint operation takes about 200 msec

FAASCONT: The container checkpoint create diff and write content in 659.375914ms seconds 
Starting point: containerd/services/tasks/local.go
	// do not commit checkpoint image if checkpoint ImagePath is passed,
	// return if checkpointImageExists is false
	if !checkpointImageExists {
		return &api.CheckpointTaskResponse{}, nil
	}

	start = time.Now()

	// write checkpoint to the content store
	tar := archive.Diff(ctx, "", image)
	cp, err := l.writeContent(ctx, images.MediaTypeContainerd1Checkpoint, image, tar)
	// close tar first after write
	if err := tar.Close(); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
Explanation: 
+Diff returns a tar stream of the computed filesystem difference between the provided directories. Produces a tar using OCI style file markers for deletions. Deleted files will be prepended with the prefix ".wh.". This style is based off AUFS whiteouts. (See https://github.com/opencontainers/mage-spec/blob/main/layer.md)


FAASCONT: The container checkpoint write container spec in 532.603713ms seconds
Starting point: containerd/services/tasks/local.go
	// write the config to the content store
	data, err := container.Spec.Marshal()
	if err != nil {
		return nil, err
	}
	spec := bytes.NewReader(data)
+	specD, err := l.writeContent(ctx, images.MediaTypeContainerd1CheckpointConfig, filepath.Joinimage, "spec"), spec)
	if err != nil {
		return nil, errdefs.ToGRPC(err)
	}
Explanation: 



Bottleneck-II: 
FAAS: The container write index took 1.041085652s seconds

Starting point: vendor/github.com/containerd/containerd/task.go in function Checkpoint()
'''
	start = time.Now()
	desc, err := t.writeIndex(ctx, &index)
	if err != nil {
		return nil, err
	}
'''