package ecommerce;

import io.grpc.Status;
import io.grpc.StatusException;

import java.util.HashMap;
import java.util.Map;
import java.util.UUID;
import productinfo.servicio.*;
public class ProductInfoImpl {

    private Map<String, Product> productMap = new HashMap<String, Product>();

    public void addProduct(Product request, io.grpc.stub.StreamObserver<ProductID> responseObserver) {
        UUID uuid = UUID.randomUUID();
        String randomUUIDString = uuid.toString();
        request = request.toBuilder().setId(randomUUIDString).build();
        
        productMap.put(randomUUIDString, request);
        
        ProductID id = ProductID.newBuilder().setValue(randomUUIDString).build();
        
        responseObserver.onNext(id);
        responseObserver.onCompleted();
    }

    public void getProduct(ProductID request, io.grpc.stub.StreamObserver<Product> responseObserver) {
        String id = request.getValue();
        if (productMap.containsKey(id)) {
            responseObserver.onNext((Product) productMap.get(id));
            responseObserver.onCompleted();
        } else {
            responseObserver.onError(new StatusException(Status.NOT_FOUND));
        }
    }
}
