package ecommerce.client;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;

import java.util.logging.Logger;
import productinfo.servicio.*;

public class ProductInfoClient {

    private static final Logger logger = Logger.getLogger(ProductInfoClient.class.getName());

    public static void main(String[] args) throws InterruptedException {
        ManagedChannel channel = ManagedChannelBuilder.forAddress("localhost", 50051)
                .usePlaintext()
                .build();

        ProductInfoGrpc.ProductInfoBlockingStub stub = ProductInfoGrpc.newBlockingStub(channel);

        ProductID productID = stub.addProduct(Product.newBuilder()
        		.setName("Samsung S10")
                .setDescription("Samsung Galaxy S10 is the latest smart phone, " +
                                "launched in February 2019")
                .setPrice(700.0f)
                .build());
        logger.info("Product ID: " + productID.getValue() + " added successfully.");

        Product product = stub.getProduct(productID);
        logger.info("Product: " + product.toString());

        channel.shutdown();
    }
}
