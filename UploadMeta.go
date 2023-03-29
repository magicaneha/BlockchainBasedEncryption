func storeDocument(encryptedDocument []byte, documentID string, accessPolicy string, awsBucketName string, hyperledgerChannelName string) error {
	// Connect to the AWS S3 bucket
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return err
	}
	s3Client := s3.New(sess)
	_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(awsBucketName),
	})
	if err != nil {
		// If the bucket already exists, ignore the error
		if !strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") {
			return err
		}
	}

	// Store the encrypted document on AWS S3
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(awsBucketName),
		Key:    aws.String(documentID),
		Body:   bytes.NewReader(encryptedDocument),
	})
	if err != nil {
		return err
	}

	// Connect to the Hyperledger Fabric network
	gateway, err := gateway.Connect(
		gateway.WithUser("user1"),
		gateway.WithClientCert("path/to/client/cert"),
		gateway.WithClientKey("path/to/client/key"),
		gateway.WithOrdererEndpoint("orderer.example.com:7050"),
		gateway.WithOrg("Org1"),
		gateway.WithChannel(hyperledgerChannelName),
	)
	if err != nil {
		return err
	}
	defer gateway.Close()

	// Get the metadata ledger contract
	network, err := gateway.GetNetwork(hyperledgerChannelName)
	if err != nil {
		return err
	}
	contract := network.GetContract("MetadataLedger")

	// Store metadata on the Hyperledger Fabric network
	_, err = contract.SubmitTransaction("CreateMetadata", documentID, accessPolicy)
	if err != nil {
		return err
	}

	return nil
}
