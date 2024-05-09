using System.ComponentModel.DataAnnotations;
 
public class GameDTO
{    
    public Guid GameGuid { get; set; }
    public String Title { get; set; }    
    public String Description { get; set; }    
    public DateTime PublishedDate { get; set; }
    public DateTime CreatedDate { get; set; }
 
}
