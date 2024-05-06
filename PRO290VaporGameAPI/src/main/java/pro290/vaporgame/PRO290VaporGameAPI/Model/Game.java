package pro290.vaporgame.PRO290VaporGameAPI.Model;

//import com.amazonaws.services.dynamodbv2.datamodeling.DynamoDBAttribute;
//import com.amazonaws.services.dynamodbv2.datamodeling.DynamoDBAutoGeneratedKey;
//import com.amazonaws.services.dynamodbv2.datamodeling.DynamoDBHashKey;
//import com.amazonaws.services.dynamodbv2.datamodeling.DynamoDBTable;
//import lombok.Data;

import java.time.LocalDate;
import java.util.ArrayList;

//@Data
//@DynamoDBTable(tableName = "Game")
public class Game {

//    @DynamoDBHashKey
//    @DynamoDBAutoGeneratedKey
    String Id;
//    @DynamoDBAttribute(attributeName = "title")
    String Title;
//    @DynamoDBAttribute(attributeName = "description")
    String Description;
//    @DynamoDBAttribute(attributeName = "author")
    String Author;
//    @DynamoDBAttribute(attributeName = "tags")
    ArrayList<String> Tags;
//    @DynamoDBAttribute(attributeName = "creationDate")
    String CreationDate;

    public Game(String title, String description, String author, ArrayList<String> tags) {
        Title = title;
        Description = description;
        Author = author;
        Tags = tags;
        CreationDate = LocalDate.now().toString();
    }

    public Game(String title, String description, String author) {
        Title = title;
        Description = description;
        Author = author;
        LocalDate.now().toString();
    }
}