using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;

public class Order
{
    [Key]
    public int OrderID { get; set; }  // Primary key

    [Required]
    public float Price { get; set; }

    public Guid UserGuid { get; set; }

    public Guid CartGuid { get; set; }

    public DateTime CreatedDate { get; set; }

    // Navigation property to OrderGame
    public virtual ICollection<OrderGame> OrderGames { get; set; }
}
