using System.ComponentModel.DataAnnotations;
 
public class Game
{
    [Key]
    public Guid GameGuid { get; set; }
 
    [Required]
    public Guid OrderGuid { get; set; }
 
     [Required]
    public Order Order { get; set; }
 
    [Required]
    public String Title { get; set; }
 
    [Required]
    public String Description { get; set; }
 
    [Required]
    public DateTime PublishedDate { get; set; }
 
    [Required]
    public DateTime CreatedDate { get; set; }
}
