using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;

public class OrderGame
{
    [Key, Column(Order = 0)]
    public int OrderID { get; set; }
    
    [Key, Column(Order = 1)]
    public int GameID { get; set; }

    // Navigation properties
    public virtual Order Order { get; set; }
    public virtual Game Game { get; set; }
}
