using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;

public class Game
{
    [Key]
    public int ID { get; set; }  // Primary key

    [Required]
    [MaxLength(100)]
    public string Title { get; set; } = string.Empty;

    public string Description { get; set; } = string.Empty;

    public string Tags { get; set; } = string.Empty;

    [Required]
    public float Price { get; set; }

    [Required]
    public DateTime Published { get; set; }

    public Guid OrderGuid { get; set; } // Add this property if Order and Game are related by OrderGuid

    // Navigation properties
    public virtual ICollection<LibraryGame> LibraryGames { get; set; }
    public virtual ICollection<OrderGame> OrderGames { get; set; }
}
